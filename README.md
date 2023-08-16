# DIScover COntacts - Scalable Mobile Private Contact Discovery 

Go / C++ / Java library implementing OPRF-based Private Set Intersection (PSI) with two-server Private Information Retrieval (PIR). 
Check out our [paper](https://ia.cr/2023/758) (published at ESORICS'23) for details.

## Code organization

The directories in this respository are:

- **contact-discovery**: OPRF-based PSI protocol with PIR for mobile private contact discovery
  - includes server and non-mobile client components
- **container**:
  - includes Containerfiles to simplify build and execution of this code base
- **cuckoo-filter**:
  - adapted Cuckoo filter implementation from [cuckoo-filter](https://github.com/linvon/cuckoo-filter)
- **mobile_psi_cpp**:
  - adapted OPRF-based PSI protocol from [mobile_psi_cpp](https://github.com/contact-discovery/mobile_psi_cpp)
- **mobile-contact-discovery**: 
  - Mobile client implementation of `contact-discovery`

## Requirements

- JAVA JNI libaries
- C++ compiler supporting C++14
- CMake v3.11 or higher
- Golang 1.19
- [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile) v0.0.0-20221110043201-43a038452099
- [Flatbuffers](https://google.github.io/flatbuffers/)
- [GRPC](https://grpc.io/)

See `mobile-contact-discovery/README.md` for additional requirements for our mobile implementation.

## Build Instructions

For our implementation, we provide the source code, as well as container files that allow you to run our benchmarks without the hassle of setting up all the dependencies.
Install and use the container management solution of your choice, e.g., podman or docker, and adapt the commands suggested in this README accordingly.

1. Get git submodules (required in `mobile_psi_cpp`)

  ```bash
  git submodule update --init # pull GSL and RELIC
  ```
  Note: When using Arch Linux or a similar distro for the build process, change file `relic/src/md/blake2.h` before building the project (https://github.com/Raptor3um/raptoreum/issues/48)
2. Build contact-discovery, OPRF and CF creator components using containerfiles or the commands used in those files. 


### Build Container
Containers can be built with 

```bash
podman build -f <ContainerFile> -t <tag>:<version> .
```

or by using the provided Makefile in `container`, e.g., `cd container && make pircontainer`.


Example: Build PIR Server Container 

```bash
podman build -f container/Containerfile.pir -t pir:1.0.0 .
```
                                   

### Build without Container


See the commands in the container files for specific details. 
When running or building Golang code from the command line the LD_LIBRARY_PATH has to include the `droidcrypto` shared library.


```bash
# Build `mobile_psi_cpp`
cd mobile_psi_cpp/
mkdir -p build
cd build
cmake ..
make -j `nproc`

# Copy resulting shared library to main project root and set $LD_LIBRARY_PATH
cp droidCrypto/libdroidcrypto.so ../../contact-discovery/libdroidcrypto.so
source .env

# Generate flatbuffers files
flatc --go --grpc -o contact-discovery contact-discovery/fbs/*.fbs

# Build non-mobile client
go build contact-discovery/cmd/client/psi_client.go

# Build server
go build contact-discovery/cmd/provider/provider_grpc.go 
```


## Test our PSI Protocol

Container logs can be read with `podman logs -f <container name>`.

### 1. Create Cuckoo filter 


Our benchmarks are based on prebuild cuckoo filters, stored in files. 
To create them build and use the `cfcreator` tool in `contact-discovery/cfcreator` or use the provided `Containerfile.cf`:


The generation and population of Cuckoo filters (CFs) is seperated from the main PSI protocol for faster development and testing. 
To create and store these CFs in files use the `cfcreator` tool in `contact-discovery/cf_creator` or the provided `Containerfile.cf`:


1. Build container using `Containerfile.cf` 
2. Run Container and `cfcreator` with the specified CF parameters.
    Specify the output folder with `-v`.

  ```bash
  # Build Container
  podman build -f container/Containerfile.cf -t cf:1.0.0 .

  # Create CF
  podman run -d --rm --name cf -v <folder for CFs>:/app/cf_files localhost/cf:1.0.0 ./app/cfcreator -cf=app/cf_files/<file name>.data -dbsize=<DB size> -prf=<PRF_TYPE> -threads=<# threads>
  ```

Example: 

  ```bash
  # Create CF
  podman run -d --rm --name cf -v `pwd`:/app/cf_files localhost/cf:1.0.0 ./app/cfcreator -cf=app/cf_files/cf.data -dbsize=14 -prf=GCLOWMC -threads=4
  ```


### 2. Run PIR Server

Our protocol requires two PIR servers.
Before running these, ensure that a CF file has been created. 

1. Build container using `Containerfile.pir`
2. Run container
  In the following command, specify the path to the CF file, the port on which the server will be available, the PRF type (`ECNR`, `GCLOWMC`, `GCAES`), the partition size and the number of workers/threads. 
  The partition size has to be a power of two and smaller than the database size. Only its exponent is required here.

  ```bash
  podman run -d --rm --name pir -v `pwd`/cf.data:/app/cf.data --network host localhost/pir:1.0.0 /app/provider_grpc -cf /app/cf.data -addr 0.0.0.0:50051 -segexp 13 -worker 4
  ```


Example: 
```bash
podman run -d --rm --name pir -v cf_GCLOWMC_31.data:/app/cf.data --network host localhost/pir:1.0.0 /app/provider_grpc -cf /app/cf.data -addr 0.0.0.0:50052 -segexp 24 -worker 4
```


### 3. Run OPRF Server

Our protocol requires one OPRF server.
The available PRF protocols include `ECNR`,`GCAES`, and `GCLOWMC`


1. Build container using `Containerfile.pir`
2. Run container

```bash
podman run -d --name oprf --network host --restart always localhost/oprf:1.0.0 app/droidCrypto/psi/oprf/oprf_server -port <port> -prf <PRF type>
```

Use the following command to automatically restart the OPRF server after each protocol execution. 
Here the status of the container is checked every 10 seconds. 

```bash
watch -n 10 podman start --all --filter restart-policy=always 
```

### 4. Mobile Client

1. Build mobile application, see `mobile-contact-discovery/README.md`
2. Deploy and run application on mobile phone
3. Enter parameters (according to the used server parameters)
4. Choose protocol to run and press the according button

Further details on the mobile implementation and functionality is described in `mobile-contact-discovery/README.md`.


## Test the PSI [KRS+19] Protocol

1. Build container using `Containerfile.psikrssw19`
2. Run container
  In the following command, specify the port on which the server will be available, the PRF type (`ECNR`, `GCLOWMC`, `GCAES`), the database size as power of two exponent.

  ```bash
  podman run --name psi --network host localhost/psi:1.0.0 /app/droidCrypto/psi/oprf/psi_server  -port <port> -prf <PRF> -dbsize <DB size>
  ```


Example: 
```bash
podman run --name psi --network host localhost/psi:1.0.0 /app/droidCrypto/psi/oprf/psi_server  -port 50051 -prf GCLOWMC -dbsize 14
```


## Disclaimer

This code is provided as an experimental implementation for testing purposes and should not be used in a productive environment. We cannot guarantee security and correctness.


## Acknowledgements

This project uses several other projects as building blocks.

- The base PSI protocol in `mobile_psi_cpp` and `mobile-contact-discovery` is based on [mobile_psi_cpp](https://github.com/contact-discovery/mobile_psi_cpp) and [mobile_psi_android](https://github.com/contact-discovery/mobile_psi_android).
- The base PIR protocol used in `contact-discovery` and `mobile-contact-discovery` is based on [Checklist](https://github.com/dimakogan/checklist). 
- The used (and adapted) Cuckoo filter implementation in `cuckoo-filter` is [Cuckoo Filter](https://github.com/linvon/cuckoo-filter).
