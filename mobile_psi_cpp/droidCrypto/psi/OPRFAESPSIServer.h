#pragma once

#include <droidCrypto/psi/PhasedPSIServer.h>
#include <droidCrypto/gc/circuits/AESCircuit.h>

namespace droidCrypto {
    class OPRFAESPSIServer : public PhasedPSIServer {
    public:
        OPRFAESPSIServer(ChannelWrapper& chan, size_t num_threads = 1);

        void Setup(std::vector<block> &elements) override;

        void Base() override;

        void Online() override;

        //void PRF(std::vector<block> &elements, std::vector<uint64_t> &elements_prf) override;

    private:
        SIMDAESCircuitPhases circ_;
    };
}

