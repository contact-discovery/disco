#include <droidCrypto/psi/PhasedPSIClient.h>
#include <droidCrypto/psi/ECNRPSIClient.h>
#include <droidCrypto/psi/OPRFAESPSIClient.h>
#include <droidCrypto/psi/OPRFLowMCPSIClient.h>
#include <droidCrypto/ChannelWrapper.h>
#include "oprf.h"



void getRandomElements(int num_elements, bool first, std::vector<droidCrypto::block> &elements) {
    droidCrypto::SecureRandom rnd;
    // First element is generated the same for client and server
    // first = true should only be set for CF Creation, not for Updates
    size_t i_start = 0;
    if  (first) {
        elements.push_back(droidCrypto::toBlock((const uint8_t*)"ffffffff88888888"));
        i_start = 1;   
    }
    for(int i = i_start; i < num_elements; i++) {
        elements.push_back(rnd.randBlock());
    }
}

void doECNR_OPRF(int num_elements, bool first, char* s_addr, int s_port, uint8_t* ptr) {
    std::vector<droidCrypto::block> elements;
    getRandomElements(num_elements, first, elements);

    droidCrypto::CSocketChannel chan(s_addr, s_port, false);
    droidCrypto::ECNRPSIClient client(chan);
    client.doOPRF(elements, ptr);
}


void doGCAES_OPRF(int num_elements, bool first, char* s_addr, int s_port, uint8_t* ptr) {
    std::vector<droidCrypto::block> elements;
    getRandomElements(num_elements, first, elements);

    droidCrypto::CSocketChannel chan(s_addr, s_port, false);
    droidCrypto::OPRFAESPSIClient client(chan);
    client.doOPRF(elements, ptr);
}

void doGCLowMC_OPRF(int num_elements, bool first, char* s_addr, int s_port, uint8_t* ptr) {
    std::vector<droidCrypto::block> elements;
    getRandomElements(num_elements, first, elements);

    droidCrypto::CSocketChannel chan(s_addr, s_port, false);
    droidCrypto::OPRFLowMCPSIClient client(chan);
    client.doOPRF(elements, ptr);
}