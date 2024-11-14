// main.cpp
#include <iostream>
#include <string>
#include <vector>
#include <thread>
#include <mutex>
#include <chrono>
#include <atomic>
#include <random>
#include <condition_variable>
#include <sstream>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <algorithm>
#include "cJSON.h"

using namespace std;

// Constants
const std::string serverAddr = "127.0.0.1";
const int serverPort = 8080;
const int maxClients = 50000;
const int numOfRooms = 100;
const std::chrono::microseconds connectInterval(1);
const std::chrono::seconds clientDuration(30);
const std::chrono::seconds stateInterval(5);
const int maxFiles = 5;
const std::chrono::milliseconds responseDeadline(10);
const int responseWindowSize = 1000;
const double maxSlowResponsePercent = 5.0;

// Precomputed messages
std::vector<std::string> helloMessages(maxClients);
std::vector<std::string> stateMessages(3);
std::vector<std::string> fileMessages(10);

// Mutex for shared resources
std::mutex responseTimesMutex;
std::mutex randomMutex;

// Random number generator
std::mt19937 rng(std::random_device{}());

// Structs for messages
struct StateMessage {
    struct Ping {
        double clientRtt = 0;
        double clientLatencyCalculation = 0;
        double latencyCalculation = 0;
    } ping;
    struct Playstate {
        bool paused = false;
        double position = 0;
        bool doSeek = false;
    } playstate;
};

struct FileMessage {
    struct File {
        double duration = 0;
        std::string name;
        int size = 0;
    } file;
};

// Function to generate random file names
std::string generateRandomFileName() {
    std::uniform_int_distribution<int> dist(0, 99999);
    int randomNum = dist(rng);
    return "file_" + std::to_string(randomNum);
}

// Function to serialize StateMessage using cJSON
std::string serializeStateMessage(const StateMessage& stateMsg) {
    cJSON* root = cJSON_CreateObject();
    cJSON* state = cJSON_AddObjectToObject(root, "State");
    cJSON* ping = cJSON_AddObjectToObject(state, "ping");
    cJSON_AddNumberToObject(ping, "clientRtt", stateMsg.ping.clientRtt);
    cJSON_AddNumberToObject(ping, "clientLatencyCalculation", stateMsg.ping.clientLatencyCalculation);
    cJSON_AddNumberToObject(ping, "latencyCalculation", stateMsg.ping.latencyCalculation);
    cJSON* playstate = cJSON_AddObjectToObject(state, "playstate");
    cJSON_AddBoolToObject(playstate, "paused", stateMsg.playstate.paused);
    cJSON_AddNumberToObject(playstate, "position", stateMsg.playstate.position);
    if (stateMsg.playstate.doSeek) {
        cJSON_AddBoolToObject(playstate, "doSeek", stateMsg.playstate.doSeek);
    }
    char* jsonString = cJSON_PrintUnformatted(root);
    std::string result(jsonString);
    cJSON_Delete(root);
    free(jsonString);
    return result + "\r\n";
}

// Function to serialize FileMessage using cJSON
std::string serializeFileMessage(const FileMessage& fileMsg) {
    cJSON* root = cJSON_CreateObject();
    cJSON* set = cJSON_AddObjectToObject(root, "Set");
    cJSON* file = cJSON_AddObjectToObject(set, "file");
    cJSON_AddNumberToObject(file, "duration", fileMsg.file.duration);
    cJSON_AddStringToObject(file, "name", fileMsg.file.name.c_str());
    cJSON_AddNumberToObject(file, "size", fileMsg.file.size);
    char* jsonString = cJSON_PrintUnformatted(root);
    std::string result(jsonString);
    cJSON_Delete(root);
    free(jsonString);
    return result + "\r\n";
}

// Initialization function to precompute messages
void init() {
    std::vector<std::thread> threads;

    // Precompute hello messages
    threads.emplace_back([]() {
        std::uniform_int_distribution<int> roomDist(0, numOfRooms - 1);
        for (int i = 0; i < maxClients; ++i) {
            std::string username = "user" + std::to_string(i);

            int roomNum;
            {
                std::lock_guard<std::mutex> lock(randomMutex);
                roomNum = roomDist(rng);
            }

            std::string room = "room" + std::to_string(roomNum);

            cJSON* root = cJSON_CreateObject();
            cJSON* hello = cJSON_AddObjectToObject(root, "Hello");
            cJSON_AddStringToObject(hello, "username", username.c_str());
            cJSON_AddStringToObject(hello, "version", "1.2.7");
            cJSON* roomObj = cJSON_AddObjectToObject(hello, "room");
            cJSON_AddStringToObject(roomObj, "name", room.c_str());
            char* jsonString = cJSON_PrintUnformatted(root);
            helloMessages[i] = std::string(jsonString) + "\r\n";
            cJSON_Delete(root);
            free(jsonString);
        }
    });

    // Precompute state messages
    threads.emplace_back([]() {
        for (int i = 0; i < 3; ++i) {
            StateMessage stateMsg;
            stateMsg.ping.clientRtt = 0;
            stateMsg.ping.clientLatencyCalculation = static_cast<double>(std::chrono::high_resolution_clock::now().time_since_epoch().count()) / 1e9;
            stateMsg.ping.latencyCalculation = stateMsg.ping.clientLatencyCalculation;
            stateMsg.playstate.paused = (i == 0);
            stateMsg.playstate.position = static_cast<double>(i * 100);
            if (i == 1) {
                stateMsg.playstate.doSeek = true;
            }
            stateMessages[i] = serializeStateMessage(stateMsg);
        }
    });

    // Precompute file messages
    threads.emplace_back([]() {
        std::uniform_real_distribution<double> durationDist(0.0, 1000.0);
        std::uniform_int_distribution<int> sizeDist(0, 999999);
        for (int i = 0; i < 10; ++i) {
            FileMessage fileMsg;
            {
                std::lock_guard<std::mutex> lock(randomMutex);
                fileMsg.file.duration = durationDist(rng);
                fileMsg.file.name = generateRandomFileName();
                fileMsg.file.size = sizeDist(rng);
            }
            fileMessages[i] = serializeFileMessage(fileMsg);
        }
    });

    // Wait for all threads to finish
    for (auto& th : threads) {
        if (th.joinable()) {
            th.join();
        }
    }
}

// Function to simulate a client
void simulateClient(int id, std::atomic<int>& concurrentClients, std::atomic<bool>& stopTest, std::vector<std::chrono::milliseconds>& responseTimes, std::mutex& responseMutex) {
    concurrentClients++;

    int sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd < 0) {
        concurrentClients--;
        return;
    }

    struct sockaddr_in serv_addr{};
    serv_addr.sin_family = AF_INET;
    serv_addr.sin_port = htons(serverPort);
    if (inet_pton(AF_INET, serverAddr.c_str(), &serv_addr.sin_addr) <= 0) {
        close(sockfd);
        concurrentClients--;
        return;
    }

    if (connect(sockfd, (struct sockaddr*)&serv_addr, sizeof(serv_addr)) < 0) {
        close(sockfd);
        concurrentClients--;
        return;
    }

    // Send Hello message
    std::string helloMsg = helloMessages[id];
    if (send(sockfd, helloMsg.c_str(), helloMsg.size(), 0) < 0) {
        close(sockfd);
        concurrentClients--;
        return;
    }

    // Start reading server responses
    std::thread readerThread([&]() {
        char buffer[1024];
        while (!stopTest.load()) {
            auto startTime = std::chrono::high_resolution_clock::now();
            ssize_t n = recv(sockfd, buffer, sizeof(buffer) - 1, 0);
            if (n <= 0) {
                break;
            }
            auto endTime = std::chrono::high_resolution_clock::now();
            std::chrono::milliseconds duration = std::chrono::duration_cast<std::chrono::milliseconds>(endTime - startTime);

            // Record response time
            {
                std::lock_guard<std::mutex> lock(responseMutex);
                responseTimes.push_back(duration);
            }
        }
    });

    // Simulate client actions
    auto startTime = std::chrono::steady_clock::now();
    bool paused = false;
    double position = 0.0;
    std::uniform_int_distribution<int> actionDist(0, 2);
    std::uniform_int_distribution<int> fileDist(0, 9);
    std::uniform_int_distribution<int> addFileDist(0, 9);

    while (std::chrono::steady_clock::now() - startTime < clientDuration && !stopTest.load()) {
        std::this_thread::sleep_for(stateInterval);

        int action = actionDist(rng);
        StateMessage stateMsg;
        stateMsg.ping.clientRtt = 0;
        stateMsg.ping.clientLatencyCalculation = static_cast<double>(std::chrono::high_resolution_clock::now().time_since_epoch().count()) / 1e9;
        stateMsg.ping.latencyCalculation = stateMsg.ping.clientLatencyCalculation;

        switch (action) {
            case 0:
                paused = !paused;
                stateMsg.playstate.paused = paused;
                stateMsg.playstate.position = position;
                break;
            case 1:
                position = std::uniform_real_distribution<double>(0.0, 1000.0)(rng);
                stateMsg.playstate.position = position;
                stateMsg.playstate.doSeek = true;
                stateMsg.playstate.paused = paused;
                break;
            case 2:
                if (!paused) {
                    position += stateInterval.count();
                }
                stateMsg.playstate.position = position;
                stateMsg.playstate.paused = paused;
                break;
        }

        std::string stateMsgStr = serializeStateMessage(stateMsg);
        if (send(sockfd, stateMsgStr.c_str(), stateMsgStr.size(), 0) < 0) {
            break;
        }

        // Randomly add a file to the room
        if (addFileDist(rng) < 2) {
            std::string fileMsg = fileMessages[fileDist(rng)];
            if (send(sockfd, fileMsg.c_str(), fileMsg.size(), 0) < 0) {
                break;
            }
        }
    }

    close(sockfd);
    if (readerThread.joinable()) {
        readerThread.join();
    }
    concurrentClients--;
}

// Function to test maximum concurrent connections
int testMaxConcurrentConnections() {
    std::atomic<int> concurrentClients{0};
    std::atomic<bool> stopTest{false};
    std::vector<std::chrono::milliseconds> responseTimes;
    std::mutex responseMutex;

    std::vector<std::thread> clientThreads;

    for (int i = 0; i < maxClients; ++i) {
        if (stopTest.load()) {
            break;
        }

        clientThreads.emplace_back(simulateClient, i, std::ref(concurrentClients), std::ref(stopTest), std::ref(responseTimes), std::ref(responseMutex));

        std::this_thread::sleep_for(connectInterval);

        // Check response times
        {
            std::lock_guard<std::mutex> lock(responseMutex);
            if (responseTimes.size() >= responseWindowSize) {
                int totalResponses = responseTimes.size();
                int slowResponses = std::count_if(responseTimes.end() - responseWindowSize, responseTimes.end(),
                    [](const std::chrono::milliseconds& d) { return d > responseDeadline; });

                double slowResponsePercent = (slowResponses * 100.0) / responseWindowSize;
                if (slowResponsePercent > maxSlowResponsePercent) {
                    std::cout << "Exceeded response time threshold at " << concurrentClients.load() << " concurrent connections\n";
                    stopTest.store(true);
                    break;
                }
            }
        }
    }

    // Wait for all client threads to finish
    for (auto& th : clientThreads) {
        if (th.joinable()) {
            th.join();
        }
    }

    return concurrentClients.load();
}

int main() {
    init();
    int maxConcurrentConnections = testMaxConcurrentConnections();
    std::cout << "Maximum concurrent connections before exceeding response threshold: " << maxConcurrentConnections << std::endl;
    return 0;
}