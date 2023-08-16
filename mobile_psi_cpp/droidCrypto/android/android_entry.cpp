#include <string>

#include <array>
#include <chrono>
#include <vector>

#include <droidCrypto/utils/Log.h>
#include <jni.h>
#define APPNAME "droidCrypto"

#include <droidCrypto/AES.h>
#include <droidCrypto/Defines.h>
#include <droidCrypto/SHA1.h>

#include <droidCrypto/BitVector.h>
#include <droidCrypto/ChannelWrapper.h>
#include <droidCrypto/SecureRandom.h>
#include <droidCrypto/gc/circuits/AESCircuit.h>
#include <droidCrypto/ot/NaorPinkas.h>
#include <droidCrypto/psi/ECNRPSIClient.h>
#include <droidCrypto/psi/OPRFAESPSIClient.h>
#include <droidCrypto/psi/OPRFLowMCPSIClient.h>
#include <droidCrypto/utils/Utils.h>

#include <droidCrypto/gc/circuits/TestCircuit.h>
#include <droidCrypto/ot/SimplestOT.h>
#include <droidCrypto/ot/TwoChooseOne/KosOtExtReceiver.h>
#include <droidCrypto/ot/TwoChooseOne/KosOtExtSender.h>
#include <cassert>
#include <thread>

using namespace droidCrypto;

enum PSI_TYPE {
  OPRF_LOWMC = 0,
  OPRF_AES,
  OPRF_ECNR,
};

void ECNR_OPRF(const char *ip, int port, int num_items) {
    droidCrypto::CSocketChannel chan(ip, port, false);
    droidCrypto::PhasedPSIClient * client = new droidCrypto::ECNRPSIClient(chan);
    uint8_t * res = (uint8_t *) malloc(num_items * 33);
    std::vector<droidCrypto::block> elements;
    droidCrypto::SecureRandom rnd;
    elements.push_back(droidCrypto::toBlock((const uint8_t *)"ffffffff88888888"));
    for (int i = 1; i < num_items; i++) {
        elements.push_back(rnd.randBlock());
    }
    client->doOPRF(elements, res);
    delete client;
}

void GCAES_OPRF(const char *ip, int port, int num_items) {

    droidCrypto::CSocketChannel chan(ip, port, false);
    droidCrypto::PhasedPSIClient * client = new droidCrypto::OPRFAESPSIClient(chan);
    uint8_t * res = (uint8_t *) malloc(num_items * 16);
    std::vector<droidCrypto::block> elements;
    droidCrypto::SecureRandom rnd;
    elements.push_back(droidCrypto::toBlock((const uint8_t *)"ffffffff88888888"));
    for (int i = 1; i < num_items; i++) {
        elements.push_back(rnd.randBlock());
    }
    client->doOPRF(elements, res);
    delete client;
}

void GCLOWMC_OPRF(const char *ip, int port, int num_items) {

    droidCrypto::CSocketChannel chan(ip, port, false);
    droidCrypto::PhasedPSIClient * client = new droidCrypto::OPRFLowMCPSIClient(chan);
    uint8_t * res = (uint8_t *) malloc(num_items * 16);
    std::vector<droidCrypto::block> elements;
    droidCrypto::SecureRandom rnd;
    elements.push_back(droidCrypto::toBlock((const uint8_t *)"ffffffff88888888"));
    for (int i = 1; i < num_items; i++) {
        elements.push_back(rnd.randBlock());
    }
    client->doOPRF(elements, res);
    delete client;
}

void ECNR_PSI(const char *ip, int port, int num_items) {
    droidCrypto::CSocketChannel chan(ip, port, false);
    droidCrypto::PhasedPSIClient * client = new droidCrypto::ECNRPSIClient(chan);
    std::vector<droidCrypto::block> elements;
    droidCrypto::SecureRandom rnd;
    elements.push_back(droidCrypto::toBlock((const uint8_t *)"ffffffff88888888"));
    for (int i = 1; i < num_items; i++) {
        elements.push_back(rnd.randBlock());
    }
    client->doPSI(elements);
    delete client;
}

void GCAES_PSI(const char *ip, int port, int num_items) {

    droidCrypto::CSocketChannel chan(ip, port, false);
    droidCrypto::PhasedPSIClient * client = new droidCrypto::OPRFAESPSIClient(chan);
    std::vector<droidCrypto::block> elements;
    droidCrypto::SecureRandom rnd;
    elements.push_back(droidCrypto::toBlock((const uint8_t *)"ffffffff88888888"));
    for (int i = 1; i < num_items; i++) {
        elements.push_back(rnd.randBlock());
    }
    client->doPSI(elements);
    delete client;
}

void GCLOWMC_PSI(const char *ip, int port, int num_items) {

    droidCrypto::CSocketChannel chan(ip, port, false);
    droidCrypto::PhasedPSIClient * client = new droidCrypto::OPRFLowMCPSIClient(chan);
    std::vector<droidCrypto::block> elements;
    droidCrypto::SecureRandom rnd;
    elements.push_back(droidCrypto::toBlock((const uint8_t *)"ffffffff88888888"));
    for (int i = 1; i < num_items; i++) {
        elements.push_back(rnd.randBlock());
    }
    client->doPSI(elements);
    delete client;
}


void testOPRF(const char *ip, int port, int num_items, PSI_TYPE type) {
  switch (type) {
    case OPRF_LOWMC:
        GCLOWMC_OPRF(ip, port, num_items);
      break;
    case OPRF_AES:
        GCAES_OPRF(ip, port, num_items);
      break;
    case OPRF_ECNR:
        ECNR_OPRF(ip, port, num_items);
      break;
    default:
      Log::e("PSI", "unrecognized psi protocol");
  }
}

void testPSI(const char *ip, int port, int num_items, PSI_TYPE type) {
    switch (type) {
        case OPRF_LOWMC:
            GCLOWMC_PSI(ip, port, num_items);
            break;
        case OPRF_AES:
            GCAES_PSI(ip, port, num_items);
            break;
        case OPRF_ECNR:
            ECNR_PSI(ip, port, num_items);
            break;
        default:
            Log::e("PSI", "unrecognized psi protocol");
    }
}

void testSpeed(const char *ip, int port) {
  std::vector<uint8_t> gigabit(1ULL << 27);
  droidCrypto::CSocketChannel chan(ip, port, false);
  Log::v("SPEED", "connected, starting test");
  chan.send(gigabit.data(), gigabit.size());
  chan.recv(gigabit.data(), gigabit.size());
  chan.recv(gigabit.data(), 1);
  chan.send(gigabit.data(), 1);
}

#define NUM_OTE (128)
void testKosOTe() {
  std::thread server([] {
    // server
    droidCrypto::CSocketChannel chan("127.0.0.1", 1233, 1);
    //        droidCrypto::NaorPinkas np;
    droidCrypto::SimplestOT np;
    droidCrypto::PRNG p = droidCrypto::PRNG::getTestPRNG();

    droidCrypto::BitVector choizes(128);  // length is in bits
    choizes.randomize(p);

    auto time1 = std::chrono::high_resolution_clock::now();
    std::array<droidCrypto::block, 128> baseOT;
    np.receive(choizes, baseOT, p, chan);
    auto time2 = std::chrono::high_resolution_clock::now();
    droidCrypto::KosOtExtSender sender;
    sender.setBaseOts(baseOT, choizes);

    std::vector<std::array<droidCrypto::block, 2>> mesBuf(NUM_OTE);
    droidCrypto::span<std::array<droidCrypto::block, 2>> mes(mesBuf.data(),
                                                             mesBuf.size());
    droidCrypto::Log::v("DOTe", "before send");
    sender.send(mes, p, chan);
    //    for(size_t a = 0; a < 10; a++) {
    //        droidCrypto::Log::v("DOTe", mesBuf[a][0]);
    //        droidCrypto::Log::v("DOTe", mesBuf[a][1]);
    //        droidCrypto::Log::v("DOTe", mesBuf[a][0]^mesBuf[a][1]);
    //        droidCrypto::Log::v("DOTe", "-----");
    //    }
    auto time3 = std::chrono::high_resolution_clock::now();
    for (size_t i = 0; i < NUM_OTE; i++) {
      chan.send(mesBuf[i][0]);
      chan.send(mesBuf[i][1]);
    }
    std::chrono::duration<double> baseOTs = time2 - time1;
    std::chrono::duration<double> OTes = time3 - time2;
    droidCrypto::Log::v("DOTe", "SENDER: BaseOTs: %fsec, OTe: %fsec", baseOTs,
                        OTes);
  });
  std::this_thread::sleep_for(std::chrono::milliseconds(20));
  // client
  droidCrypto::CSocketChannel chan("127.0.0.1", 1233, 0);
  //    droidCrypto::NaorPinkas np;
  droidCrypto::SimplestOT np;
  droidCrypto::PRNG p = droidCrypto::PRNG::getTestPRNG();

  auto time1 = std::chrono::high_resolution_clock::now();
  std::array<std::array<droidCrypto::block, 2>, 128> baseOT;
  np.send(baseOT, p, chan);
  auto time2 = std::chrono::high_resolution_clock::now();
  droidCrypto::KosOtExtReceiver recv;
  recv.setBaseOts(baseOT);

  droidCrypto::BitVector choizes(NUM_OTE);
  choizes.randomize(p);
  std::vector<droidCrypto::block> mesBuf(NUM_OTE);
  droidCrypto::span<droidCrypto::block> mes(mesBuf.data(), mesBuf.size());
  droidCrypto::Log::v("DOTe", "before recv");
  recv.receive(choizes, mes, p, chan);
  auto time3 = std::chrono::high_resolution_clock::now();
  std::vector<std::array<droidCrypto::block, 2>> buf(NUM_OTE);
  for (size_t i = 0; i < NUM_OTE; i++) {
    chan.recv(buf[i][0]);
    chan.recv(buf[i][1]);
    if (droidCrypto::neq(buf[i][uint8_t(choizes[i])], mesBuf[i])) {
      droidCrypto::Log::e("DOTe", "Wrong in %d (%d)!", i, uint8_t(choizes[i]));
      droidCrypto::Log::v("DOTe", buf[i][0]);
      droidCrypto::Log::v("DOTe", buf[i][1]);
      droidCrypto::Log::v("DOTe", mesBuf[i]);
      droidCrypto::Log::v("DOTe", "----------------");
    }
  }

  std::chrono::duration<double> baseOTs = time2 - time1;
  std::chrono::duration<double> OTes = time3 - time2;
  droidCrypto::Log::v("DOTe", "RECVER: BaseOTs: %fsec, OTe: %fsec", baseOTs,
                      OTes);
  droidCrypto::Log::v("DOTe", "RECVER: recv: %zu, sent: %zu",
                      chan.getBytesRecv(), chan.getBytesSent());
  server.join();
}

using namespace droidCrypto;

extern "C" JNIEXPORT void JNICALL
Java_com_example_mobile_1psi_droidCrypto_TestSpeedTask_testSpeed(
    JNIEnv *env, jobject /* this */,jstring jip, jint port) {
    const jsize iplen = env->GetStringUTFLength(jip);
    const char *ip = env->GetStringUTFChars(jip, (jboolean *)0);
    testSpeed(ip, port);
}
extern "C" JNIEXPORT jstring JNICALL
Java_com_example_mobile_1psi_droidCrypto_TestAsyncTask_testNative(
    JNIEnv *env, jclass clazz, jstring jip, jlong port, jlong type,
    jlong num_items) {
  auto start = std::chrono::high_resolution_clock::now();
  const jsize iplen = env->GetStringUTFLength(jip);
  const char *ip = env->GetStringUTFChars(jip, (jboolean *)0);

  Log::v("PSI", "connecting to %s:%d", ip, port);
  testOPRF(ip, port, num_items, PSI_TYPE(type));
  env->ReleaseStringUTFChars(jip, ip);

  auto finish = std::chrono::high_resolution_clock::now();
  std::chrono::duration<double> elapsed = finish - start;
  std::string time_taken = "Time: " + std::to_string(elapsed.count());
  return env->NewStringUTF(time_taken.c_str());
}


extern "C" JNIEXPORT void JNICALL
Java_com_example_mobile_1psi_droidCrypto_TestAsyncTask_testSend(
    JNIEnv *env, jobject thisObj, jobject channel) {
  uint8_t test[] = {0x12, 0x34, 0x56, 0x78};

  JavaChannelWrapper cw(env, channel);
  cw.send(test, 4);

  Log::v(APPNAME, "sent: %02X%02X%02X%02X", test[0], test[1], test[2], test[3]);
}

extern "C" JNIEXPORT void JNICALL
Java_com_example_mobile_1psi_droidCrypto_TestAsyncTask_testRecv(
    JNIEnv *env, jobject thisObj, jobject channel) {
  uint8_t test[4] = {
      0,
  };

  JavaChannelWrapper cw(env, channel);
  cw.recv(test, 4);
  Log::v(APPNAME, "recvd: %02X%02X%02X%02X", test[0], test[1], test[2],
         test[3]);
}

extern "C"
JNIEXPORT jstring JNICALL
Java_com_example_mobile_1psi_droidCrypto_TestPSIOldAsyncTask_testNative(
        JNIEnv *env, jobject thiz,
        jstring jip, jint port,
        jlong type,
        jint num_items) {
        auto start = std::chrono::high_resolution_clock::now();
        const jsize iplen = env->GetStringUTFLength(jip);
        const char *ip = env->GetStringUTFChars(jip, (jboolean *)0);

        Log::v("PSI", "connecting to %s:%d", ip, port);
        testPSI(ip, port, num_items, PSI_TYPE(type));
        env->ReleaseStringUTFChars(jip, ip);

        auto finish = std::chrono::high_resolution_clock::now();
        std::chrono::duration<double> elapsed = finish - start;
        std::string time_taken = "Time: " + std::to_string(elapsed.count());
        return env->NewStringUTF(time_taken.c_str());
}