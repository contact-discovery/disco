
#include <droidCrypto/psi/PhasedPSIServer.h>
#include <droidCrypto/ChannelWrapper.h>
#include <droidCrypto/psi/OPRFAESPSIServer.h>
#include <droidCrypto/psi/OPRFLowMCPSIServer.h>
#include <droidCrypto/psi/ECNRPSIServer.h>
#include <thread>
#include <iostream>

//namespace oprf{

int main(int argc, char** argv) {

    std::string arg_port_str("-port");
    std::string arg_prf_str("-prf");

    int port = 50051;
    std::string prf_type = "ECNR";

    for (int i=1; i < argc; i+=2) {
        if(argv[i] == arg_port_str) {
            port = std::stoi(argv[i+1]);
        } else if(argv[i] == arg_prf_str) {
            prf_type = argv[i+1];
            if ((prf_type != "ECNR") && (prf_type != "GCAES") && (prf_type != "GCLOWMC")) {
                std::cout << "The correct argument syntax is -port <PORT> -prf <ECNR|GCAES|GCLOWMC>"
                    << std::endl;
                return 0;
            }
        }
    }
    if (prf_type == "ECNR") {      
        std::cout << "Start ECNR-OPRF Server on port " << port <<"\n";
        droidCrypto::CSocketChannel chan(nullptr, port, true);
        droidCrypto::ECNRPSIServer server(chan, 1);
        server.doOPRF();
        std::cout << "Done ECNR-OPRF\n";
    } else if (prf_type == "GCAES") {
        std::cout << "Start GCAES-OPRF Server on port " << port <<"\n";  
        droidCrypto::CSocketChannel chan(nullptr, port, true);
        droidCrypto::OPRFAESPSIServer server(chan, 1);
        server.doOPRF();
        std::cout << "Done GCAES-OPRF\n";
    } else if (prf_type == "GCLOWMC") {
        std::cout << "Start GCLowMC-OPRF Server on port " << port <<"\n"; 
        droidCrypto::CSocketChannel chan(nullptr, port, true);
        droidCrypto::OPRFLowMCPSIServer server(chan, 1);
        server.doOPRF();
        std::cout << "Done GCLowMC-OPRF\n";
    }
    return 1;
}

