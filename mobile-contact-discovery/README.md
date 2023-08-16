# Mobile Private Contact Discovery

Android application for our Mobile Private Contact Discovery protocol. 

This application is based on the following libraries:
- [Mobile Private Contact Discovery](https://github.com/contact-discovery/mobile_psi_android) 
- [Checklist](https://github.com/dimakogan/checklist)


## Requirements

- Phone with ARM64-v8a ABI
- JAVA JNI libaries
- C++ compiler supporting C++14
- CMake v3.11 or higher
- Golang 1.19
- [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile)  v0.0.0-20221110043201-43a038452099
* [Flatbuffers](https://google.github.io/flatbuffers/)
* [GRPC](https://grpc.io/)


The application was built using Android Studio with the following project settings:
    - Gradle JDK 1.8 
    - Android Gradle Plugin Version 4.2.0
    - Gradle Version 7.4
    - Compile Sdk Version 29
    - Min Sdk Version 17
    

## Build App
Add the required code to this directory

   ```bash
   cp -r ../contact-discovery/pir/ ./ && \
   cp -r ../contact-discovery/psetggm/ ./ && \
   cp -r ../contact-discovery/fbs/ ./ && \
   cp -r ../mobile_psi_cpp/ ./android/app/src/main/cpp/
   ```
   If the Flatbuffers files have not already been built in `contact-discovery` do so in this directory:
   ```bash
   flatc --go --grpc -o ./  fbs/*.fbs
   ```

2. Install gomobile
    ```bash
    go install golang.org/x/mobile/cmd/gomobile@v0.0.0-20221110043201-43a038452099
    gomobile init
    ```
3. Generate the Java bindings
    ```bash
    go get -d golang.org/x/mobile/cmd/gomobile
    gomobile bind -o android/app/pir.aar -target android/arm64 .
    ```

4. Build and deploy the application, e.g., using Android Studio.

## Test Application

### Our PSI protocol

1. Start OPRF server and two PIR server
2. Fill out form in our mobile application
   1. Server IPs and ports (*PIR1*, *PIR2*, *OPRF*)
   2. Server and client set size and partition size. All set sizes are powers of two.
   3. Mapping threshold
   4. PRF
   5. Number of workers
3. Press Button *PSI*
4. Wait
5. Results will be shown after protocol execution

The *OPRF* protocol and our PIR-based PSI (*PIR*) can be run independently using the according buttons. 


### PSI [KRS+19] protocol


1. Start PSI (KRS+19) server
2. Fill out form in our mobile application
   1. Server IP and port (*OPRF*)
   2. Server and client set size. All set sizes are powers of two.
   3. PRF
3. Press Button "PSI [KRS+19]"
4. Wait
5. Results will be shown after protocol execution



### Partition Test

Simulates the execution of our PSI protocol in a distributed setting, where each database partition is held by two (online and offline) servers. 
In this test, one pair of servers processes one database partition, but sends multiple copies (one copy for each of the database partitions) to the client who processes all. 
This allows the simulation of server and client performance in a distributed setting.

Fill out the form for our PSI protocol and press *Partition Test*
