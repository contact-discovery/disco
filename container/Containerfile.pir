FROM docker.io/golang:1.19 as build
RUN apt-get update && \
    apt-get install -y curl unzip build-essential cmake autoconf automake libtool libboost-all-dev make git g++ openssl libssl-dev openjdk-17-jdk
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64/
WORKDIR /root/go/src/disco
ADD ./mobile_psi_cpp ./mobile_psi_cpp
ADD ./contact-discovery ./contact-discovery
ADD ./mobile-contact-discovery ./mobile-contact-discovery
ADD ./cuckoo-filter ./cuckoo-filter
ADD ./go.work* ./

RUN mkdir -p mobile_psi_cpp/build && \
    cd mobile_psi_cpp/build && \
    cmake .. && \
    make -j `nproc`

RUN curl -L -o /tmp/flatc.zip https://github.com/google/flatbuffers/releases/download/v22.10.26/Linux.flatc.binary.g++-10.zip && \
    unzip -d /tmp/ /tmp/flatc.zip && \
    chmod +x /tmp/flatc && \
    mv /tmp/flatc /usr/local/bin

RUN cp /root/go/src/disco/mobile_psi_cpp/build/droidCrypto/keccak/libkeccak.so /root/go/src/disco/contact-discovery/libkeccak.so
RUN cp /root/go/src/disco/mobile_psi_cpp/build/droidCrypto/libdroidcrypto.so /root/go/src/disco/contact-discovery/libdroidcrypto.so
ENV LD_LIBRARY_PATH /root/go/src/disco/contact-discovery

RUN flatc --go --grpc -o /root/go/src/disco/contact-discovery /root/go/src/disco/contact-discovery/fbs/*.fbs

RUN cd /root/go/src/disco/contact-discovery && go get -d ./...
RUN go build /root/go/src/disco/contact-discovery/cmd/provider/provider_grpc.go

FROM docker.io/ubuntu:22.04
COPY --from=build /root/go/src/disco/contact-discovery/libkeccak.so /app/libs/libkeccak.so
COPY --from=build /root/go/src/disco/contact-discovery/libdroidcrypto.so /app/libs/libdroidcrypto.so
ENV LD_LIBRARY_PATH /app/libs
COPY --from=build /root/go/src/disco/provider_grpc /app/
cmd /app/provider_grpc
