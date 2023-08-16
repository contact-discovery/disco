#include <droidCrypto/SecureRandom.h>
#include <droidCrypto/psi/tools/ECNRPRF.h>
#include <droidCrypto/gc/circuits/LowMCCircuit.h>
#include <thread>
#include <droidCrypto/utils/Log.h>
#include <droidCrypto/Defines.h>
#include "prf.h"
#include <iostream>


extern "C" {
    #include <droidCrypto/lowmc/lowmc_pars.h>
    #include <droidCrypto/lowmc/io.h>
    #include <droidCrypto/lowmc/lowmc.h>
    #include <droidCrypto/lowmc/lowmc_128_128_192.h>
}

using droidCrypto::block;
using droidCrypto::Log;


void getRandomElementsPRF(int num_elements, bool first, std::vector<droidCrypto::block> &elements) {

    droidCrypto::SecureRandom rnd;
    // First element is generated the same for client and server
    // first = true should only be set for CF Creation, not for Updates
    size_t i_start = 0;
    if  (first) {
        elements.push_back(droidCrypto::toBlock((const uint8_t*)"ffffffff88888888"));
        i_start = 1;   
    }
    for(size_t i = i_start; i < num_elements; i++) {
        elements.push_back(rnd.randBlock());
    }
}

void getECNR_PRF(int num_elements, bool first, int num_threads, uint8_t* ptr) {
    std::vector<droidCrypto::block> elements;
    getRandomElementsPRF(num_elements, first, elements);
    std::cerr << "got random elements\n";
    // MT-bounds
    size_t elements_per_thread = num_elements / num_threads;
    Log::v("PRF", "%zu threads, %zu elements each", num_threads, elements_per_thread);
    std::cerr << " ";
    droidCrypto::PRNG prng(droidCrypto::PRNG::getTestPRNG());
    droidCrypto::ECNRPRF prf(prng, 128); 

    for (size_t i = 0; i < num_elements; i++) {
        prf.prf(elements[i]).toBytes(ptr);
        ptr = ptr+33;
    }
    std::cerr << "did prf\n";
}


void getGCAES_PRF(int num_elements, bool first, int num_threads, uint8_t* ptr) {
    std::vector<droidCrypto::block> elements;
    getRandomElementsPRF(num_elements, first, elements);

    std::cout << "GCAES PRF - rand elements done\n";
  // MT-bounds
    size_t elements_per_thread = num_elements / num_threads;
    Log::v("PSI", "%zu threads, %zu elements each", num_threads, elements_per_thread);
    uint8_t AES_TEST_KEY[16] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
    droidCrypto::AES a;
    a.setKey(AES_TEST_KEY);
    std::vector<std::thread> threads;
    for (size_t thrd = 0; thrd < num_threads - 1; thrd++) {
        auto t = std::thread([aes = a, &elements, elements_per_thread, idx = thrd] {
        aes.encryptECBBlocks(elements.data() + idx * elements_per_thread,
                           elements_per_thread,
                           elements.data() + idx * elements_per_thread);
        });
        threads.emplace_back(std::move(t));
    }
    // rest in main thread
    a.encryptECBBlocks(
        elements.data() + (num_threads - 1) * elements_per_thread,
        num_elements - (num_threads - 1) * elements_per_thread,
        elements.data() + (num_threads - 1) * elements_per_thread);
    for (size_t thrd = 0; thrd < num_threads - 1; thrd++) {
        threads[thrd].join();
    }

    for (size_t i = 0; i < num_elements; i++) {
        memcpy(ptr, &elements[i], 16);
        ptr = ptr+16;
    }
    std::cout << "GCAES PRF done\n";
}

void getGCLowMC_PRF(int num_elements, bool first, int num_threads, uint8_t* ptr) {
    std::vector<droidCrypto::block> elements;
    getRandomElementsPRF(num_elements, first, elements);

    //MT-bounds
    size_t elements_per_thread = num_elements / num_threads;
    Log::v("PSI", "%zu threads, %zu elements each", num_threads, elements_per_thread);
    //LOWMC encryption
    // get a random key
    std::array<uint8_t, 16> lowmc_key;
    droidCrypto::PRNG::getTestPRNG().get(lowmc_key.data(), lowmc_key.size());

    const lowmc_t* params = droidCrypto::SIMDLowMCCircuitPhases::params;
    lowmc_key_t* key = mzd_local_init(1, params->k);
    mzd_from_char_array(key, lowmc_key.data(), (params->k)/8);
    expanded_key key_calc = lowmc_expand_key(params, key);

    std::vector<std::thread> threads;
    for(size_t thrd = 0; thrd < num_threads-1; thrd++) {
        auto t = std::thread([params, key_calc, &elements, elements_per_thread,idx=thrd]{
            lowmc_key_t* pt = mzd_local_init(1, params->n);
            for(size_t i = idx*elements_per_thread; i < (idx+1)*elements_per_thread; i++) {
                mzd_from_char_array(pt, (uint8_t *) (&elements[i]), params->n / 8);
                mzd_local_t *ct = lowmc_call(params, key_calc, pt);
                mzd_to_char_array((uint8_t *) (&elements[i]), ct, params->n / 8);
                mzd_local_free(ct);
            }
            mzd_local_free(pt);
        });
        threads.emplace_back(std::move(t));
    }
    lowmc_key_t* pt = mzd_local_init(1, params->n);
    for(size_t i = (num_threads-1)*elements_per_thread; i < num_elements; i++) {
        mzd_from_char_array(pt, (uint8_t *) (&elements[i]), params->n / 8);
        mzd_local_t *ct = lowmc_call(params, key_calc, pt);
        mzd_to_char_array((uint8_t *) (&elements[i]), ct, (params->n) / 8);
        mzd_local_free(ct);
    }
    mzd_local_free(pt);
    for(size_t thrd = 0; thrd < num_threads -1; thrd++) {
        threads[thrd].join();
    }
    for (size_t i = 0; i < num_elements; i++) {
        memcpy(ptr, &elements[i], 16);
        ptr = ptr+16;
    }
    std::cout << "GCLowMC PRF done\n";
}