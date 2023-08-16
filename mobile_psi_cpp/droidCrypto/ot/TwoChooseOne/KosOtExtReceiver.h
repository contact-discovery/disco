#pragma once
// This file and the associated implementation has been placed in the public domain, waiving all copyright. No restrictions are placed on its use. 
#include "OTExtInterface.h"
#include <array>
#include <droidCrypto/PRNG.h>


namespace droidCrypto
{

    class KosOtExtReceiver :
        public OtExtReceiver
    {
    public:
        KosOtExtReceiver()
            :mHasBase(false)
        {}

        bool hasBaseOts() const override
        {
            return mHasBase;
        }

        bool mHasBase;
        std::array<std::array<PRNG, 2>, gOtExtBaseOtCount> mGens;

        void setBaseOts(
            span<std::array<block, 2>> baseSendOts)override;


        std::unique_ptr<OtExtReceiver> split() override;

        void receive(
            const BitVector& choices,
            span<block> messages,
            PRNG& prng,
            ChannelWrapper& chl/*,
            std::atomic<u64>& doneIdx*/) override;


    };

}
