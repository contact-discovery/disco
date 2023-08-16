/*
* SHA-1 hash in C
*
* Copyright (c) 2014 Project Nayuki
* https://www.nayuki.io/page/fast-sha1-hash-implementation-in-x86-assembly
*
* (MIT License)
* Permission is hereby granted, free of charge, to any person obtaining a copy of
* this software and associated documentation files (the "Software"), to deal in
* the Software without restriction, including without limitation the rights to
* use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
* the Software, and to permit persons to whom the Software is furnished to do so,
* subject to the following conditions:
* - The above copyright notice and this permission notice shall be included in
*   all copies or substantial portions of the Software.
* - The Software is provided "as is", without warranty of any kind, express or
*   implied, including but not limited to the warranties of merchantability,
*   fitness for a particular purpose and noninfringement. In no event shall the
*   authors or copyright holders be liable for any claim, damages or other
*   liability, whether in an action of contract, tort or otherwise, arising from,
*   out of or in connection with the Software or the use or other dealings in the
*   Software.
*/
#include <droidCrypto/SHA1.h>
#include <droidCrypto/Defines.h>
#include <stdint.h>
#include <string>
#include <cstring>



#if defined(HAVE_NEON)
//TODO: FIX SHA-1
#define NO_ARM_NEON_SHA1
#else
#define NO_ARM_NEON_SHA1
#endif

void sha1_compress(uint32_t state[5], const uint8_t block[64])
{

#ifndef NO_ARM_NEON_SHA1

    // disable this if you dont want the assembly version.
    const uint32_t* data = reinterpret_cast<const uint32_t*>(&block[0]);

    uint32x4_t C0, C1, C2, C3;
    uint32x4_t ABCD, ABCD_SAVED;
    uint32x4_t MSG0, MSG1, MSG2, MSG3;
    uint32x4_t TMP0, TMP1;
    uint32_t   E0, E0_SAVED, E1;

    // Load initial values
    C0 = vdupq_n_u32(0x5A827999);
    C1 = vdupq_n_u32(0x6ED9EBA1);
    C2 = vdupq_n_u32(0x8F1BBCDC);
    C3 = vdupq_n_u32(0xCA62C1D6);

    ABCD = vld1q_u32(&state[0]);
    E0 = state[4];

    // Save current hash
    ABCD_SAVED = ABCD;
    E0_SAVED = E0;

    MSG0 = vld1q_u32(data +  0);
    MSG1 = vld1q_u32(data +  4);
    MSG2 = vld1q_u32(data +  8);
    MSG3 = vld1q_u32(data + 12);

    TMP0 = vaddq_u32(MSG0, C0);
    TMP1 = vaddq_u32(MSG1, C0);

    // Rounds 0-3
    E1 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1cq_u32(ABCD, E0, TMP0);
    TMP0 = vaddq_u32(MSG2, C0);
    MSG0 = vsha1su0q_u32(MSG0, MSG1, MSG2);

    // Rounds 4-7
    E0 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1cq_u32(ABCD, E1, TMP1);
    TMP1 = vaddq_u32(MSG3, C0);
    MSG0 = vsha1su1q_u32(MSG0, MSG3);
    MSG1 = vsha1su0q_u32(MSG1, MSG2, MSG3);

    // Rounds 8-11
    E1 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1cq_u32(ABCD, E0, TMP0); /* 2 */
    TMP0 = vaddq_u32(MSG0, C0);
    MSG1 = vsha1su1q_u32(MSG1, MSG0);
    MSG2 = vsha1su0q_u32(MSG2, MSG3, MSG0);

    // Rounds 12-15
    E0 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1cq_u32(ABCD, E1, TMP1);
    TMP1 = vaddq_u32(MSG1, C1);
    MSG2 = vsha1su1q_u32(MSG2, MSG1);
    MSG3 = vsha1su0q_u32(MSG3, MSG0, MSG1);

    // Rounds 16-19
    E1 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1cq_u32(ABCD, E0, TMP0);
    TMP0 = vaddq_u32(MSG2, C1);
    MSG3 = vsha1su1q_u32(MSG3, MSG2);
    MSG0 = vsha1su0q_u32(MSG0, MSG1, MSG2);

    // Rounds 20-23
    E0 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1pq_u32(ABCD, E1, TMP1);
    TMP1 = vaddq_u32(MSG3, C1);
    MSG0 = vsha1su1q_u32(MSG0, MSG3);
    MSG1 = vsha1su0q_u32(MSG1, MSG2, MSG3);

    // Rounds 24-27
    E1 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1pq_u32(ABCD, E0, TMP0);
    TMP0 = vaddq_u32(MSG0, C1);
    MSG1 = vsha1su1q_u32(MSG1, MSG0);
    MSG2 = vsha1su0q_u32(MSG2, MSG3, MSG0);

    // Rounds 28-31
    E0 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1pq_u32(ABCD, E1, TMP1);
    TMP1 = vaddq_u32(MSG1, C1);
    MSG2 = vsha1su1q_u32(MSG2, MSG1);
    MSG3 = vsha1su0q_u32(MSG3, MSG0, MSG1);

    // Rounds 32-35
    E1 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1pq_u32(ABCD, E0, TMP0);
    TMP0 = vaddq_u32(MSG2, C2);
    MSG3 = vsha1su1q_u32(MSG3, MSG2);
    MSG0 = vsha1su0q_u32(MSG0, MSG1, MSG2);

    // Rounds 36-39
    E0 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1pq_u32(ABCD, E1, TMP1);
    TMP1 = vaddq_u32(MSG3, C2);
    MSG0 = vsha1su1q_u32(MSG0, MSG3);
    MSG1 = vsha1su0q_u32(MSG1, MSG2, MSG3);

    // Rounds 40-43
    E1 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1mq_u32(ABCD, E0, TMP0);
    TMP0 = vaddq_u32(MSG0, C2);
    MSG1 = vsha1su1q_u32(MSG1, MSG0);
    MSG2 = vsha1su0q_u32(MSG2, MSG3, MSG0);

    // Rounds 44-47
    E0 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1mq_u32(ABCD, E1, TMP1);
    TMP1 = vaddq_u32(MSG1, C2);
    MSG2 = vsha1su1q_u32(MSG2, MSG1);
    MSG3 = vsha1su0q_u32(MSG3, MSG0, MSG1);

    // Rounds 48-51
    E1 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1mq_u32(ABCD, E0, TMP0);
    TMP0 = vaddq_u32(MSG2, C2);
    MSG3 = vsha1su1q_u32(MSG3, MSG2);
    MSG0 = vsha1su0q_u32(MSG0, MSG1, MSG2);

    // Rounds 52-55
    E0 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1mq_u32(ABCD, E1, TMP1);
    TMP1 = vaddq_u32(MSG3, C3);
    MSG0 = vsha1su1q_u32(MSG0, MSG3);
    MSG1 = vsha1su0q_u32(MSG1, MSG2, MSG3);

    // Rounds 56-59
    E1 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1mq_u32(ABCD, E0, TMP0);
    TMP0 = vaddq_u32(MSG0, C3);
    MSG1 = vsha1su1q_u32(MSG1, MSG0);
    MSG2 = vsha1su0q_u32(MSG2, MSG3, MSG0);

    // Rounds 60-63
    E0 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1pq_u32(ABCD, E1, TMP1);
    TMP1 = vaddq_u32(MSG1, C3);
    MSG2 = vsha1su1q_u32(MSG2, MSG1);
    MSG3 = vsha1su0q_u32(MSG3, MSG0, MSG1);

    // Rounds 64-67
    E1 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1pq_u32(ABCD, E0, TMP0);
    TMP0 = vaddq_u32(MSG2, C3);
    MSG3 = vsha1su1q_u32(MSG3, MSG2);
    MSG0 = vsha1su0q_u32(MSG0, MSG1, MSG2);

    // Rounds 68-71
    E0 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1pq_u32(ABCD, E1, TMP1);
    TMP1 = vaddq_u32(MSG3, C3);
    MSG0 = vsha1su1q_u32(MSG0, MSG3);

    // Rounds 72-75
    E1 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1pq_u32(ABCD, E0, TMP0);

    // Rounds 76-79
    E0 = vsha1h_u32(vgetq_lane_u32(ABCD, 0));
    ABCD = vsha1pq_u32(ABCD, E1, TMP1);

    E0 += E0_SAVED;
    ABCD = vaddq_u32(ABCD_SAVED, ABCD);

    // Save state
    vst1q_u32(&state[0], ABCD);
    state[4] = E0;


#else
#define ROTL32(x, n)  (((0U + (x)) << (n)) | ((x) >> (32 - (n))))  // Assumes that x is uint32_t and 0 < n < 32

#define LOADSCHEDULE(i)  \
        schedule[i] = (uint32_t)block[i * 4 + 0] << 24  \
                    | (uint32_t)block[i * 4 + 1] << 16  \
                    | (uint32_t)block[i * 4 + 2] <<  8  \
                    | (uint32_t)block[i * 4 + 3] <<  0;

#define SCHEDULE(i)  \
        temp = schedule[(i - 3) & 0xF] ^ schedule[(i - 8) & 0xF] ^ schedule[(i - 14) & 0xF] ^ schedule[(i - 16) & 0xF];  \
        schedule[i & 0xF] = ROTL32(temp, 1);

#define ROUND0a(a, b, c, d, e, i)  LOADSCHEDULE(i)  ROUNDTAIL(a, b, e, ((b & c) | (~b & d))         , i, 0x5A827999)
#define ROUND0b(a, b, c, d, e, i)  SCHEDULE(i)      ROUNDTAIL(a, b, e, ((b & c) | (~b & d))         , i, 0x5A827999)
#define ROUND1(a, b, c, d, e, i)   SCHEDULE(i)      ROUNDTAIL(a, b, e, (b ^ c ^ d)                  , i, 0x6ED9EBA1)
#define ROUND2(a, b, c, d, e, i)   SCHEDULE(i)      ROUNDTAIL(a, b, e, ((b & c) ^ (b & d) ^ (c & d)), i, 0x8F1BBCDC)
#define ROUND3(a, b, c, d, e, i)   SCHEDULE(i)      ROUNDTAIL(a, b, e, (b ^ c ^ d)                  , i, 0xCA62C1D6)

#define ROUNDTAIL(a, b, e, f, i, k)  \
        e = 0U + e + ROTL32(a, 5) + f + UINT32_C(k) + schedule[i & 0xF];  \
        b = ROTL32(b, 30);

    uint32_t a = state[0];
    uint32_t b = state[1];
    uint32_t c = state[2];
    uint32_t d = state[3];
    uint32_t e = state[4];

    uint32_t schedule[16];
    uint32_t temp;
    ROUND0a(a, b, c, d, e, 0)
        ROUND0a(e, a, b, c, d, 1)
        ROUND0a(d, e, a, b, c, 2)
        ROUND0a(c, d, e, a, b, 3)
        ROUND0a(b, c, d, e, a, 4)
        ROUND0a(a, b, c, d, e, 5)
        ROUND0a(e, a, b, c, d, 6)
        ROUND0a(d, e, a, b, c, 7)
        ROUND0a(c, d, e, a, b, 8)
        ROUND0a(b, c, d, e, a, 9)
        ROUND0a(a, b, c, d, e, 10)
        ROUND0a(e, a, b, c, d, 11)
        ROUND0a(d, e, a, b, c, 12)
        ROUND0a(c, d, e, a, b, 13)
        ROUND0a(b, c, d, e, a, 14)
        ROUND0a(a, b, c, d, e, 15)
        ROUND0b(e, a, b, c, d, 16)
        ROUND0b(d, e, a, b, c, 17)
        ROUND0b(c, d, e, a, b, 18)
        ROUND0b(b, c, d, e, a, 19)
        ROUND1(a, b, c, d, e, 20)
        ROUND1(e, a, b, c, d, 21)
        ROUND1(d, e, a, b, c, 22)
        ROUND1(c, d, e, a, b, 23)
        ROUND1(b, c, d, e, a, 24)
        ROUND1(a, b, c, d, e, 25)
        ROUND1(e, a, b, c, d, 26)
        ROUND1(d, e, a, b, c, 27)
        ROUND1(c, d, e, a, b, 28)
        ROUND1(b, c, d, e, a, 29)
        ROUND1(a, b, c, d, e, 30)
        ROUND1(e, a, b, c, d, 31)
        ROUND1(d, e, a, b, c, 32)
        ROUND1(c, d, e, a, b, 33)
        ROUND1(b, c, d, e, a, 34)
        ROUND1(a, b, c, d, e, 35)
        ROUND1(e, a, b, c, d, 36)
        ROUND1(d, e, a, b, c, 37)
        ROUND1(c, d, e, a, b, 38)
        ROUND1(b, c, d, e, a, 39)
        ROUND2(a, b, c, d, e, 40)
        ROUND2(e, a, b, c, d, 41)
        ROUND2(d, e, a, b, c, 42)
        ROUND2(c, d, e, a, b, 43)
        ROUND2(b, c, d, e, a, 44)
        ROUND2(a, b, c, d, e, 45)
        ROUND2(e, a, b, c, d, 46)
        ROUND2(d, e, a, b, c, 47)
        ROUND2(c, d, e, a, b, 48)
        ROUND2(b, c, d, e, a, 49)
        ROUND2(a, b, c, d, e, 50)
        ROUND2(e, a, b, c, d, 51)
        ROUND2(d, e, a, b, c, 52)
        ROUND2(c, d, e, a, b, 53)
        ROUND2(b, c, d, e, a, 54)
        ROUND2(a, b, c, d, e, 55)
        ROUND2(e, a, b, c, d, 56)
        ROUND2(d, e, a, b, c, 57)
        ROUND2(c, d, e, a, b, 58)
        ROUND2(b, c, d, e, a, 59)
        ROUND3(a, b, c, d, e, 60)
        ROUND3(e, a, b, c, d, 61)
        ROUND3(d, e, a, b, c, 62)
        ROUND3(c, d, e, a, b, 63)
        ROUND3(b, c, d, e, a, 64)
        ROUND3(a, b, c, d, e, 65)
        ROUND3(e, a, b, c, d, 66)
        ROUND3(d, e, a, b, c, 67)
        ROUND3(c, d, e, a, b, 68)
        ROUND3(b, c, d, e, a, 69)
        ROUND3(a, b, c, d, e, 70)
        ROUND3(e, a, b, c, d, 71)
        ROUND3(d, e, a, b, c, 72)
        ROUND3(c, d, e, a, b, 73)
        ROUND3(b, c, d, e, a, 74)
        ROUND3(a, b, c, d, e, 75)
        ROUND3(e, a, b, c, d, 76)
        ROUND3(d, e, a, b, c, 77)
        ROUND3(c, d, e, a, b, 78)
        ROUND3(b, c, d, e, a, 79)

        state[0] = 0U + state[0] + a;
    state[1] = 0U + state[1] + b;
    state[2] = 0U + state[2] + c;
    state[3] = 0U + state[3] + d;
    state[4] = 0U + state[4] + e;
#endif


    //if (memcmp(state, state2, sizeof(uint32_t) * 5))
    //{
    //    throw std::runtime_error("a");
    //}

    //if (memcmp(block, block2, sizeof(uint8_t) * 64))
    //{
    //    throw std::runtime_error("b");
    //}
}
namespace droidCrypto
{
    const uint64_t    SHA1::HashSize;


    const SHA1& SHA1::operator=(const SHA1& src)
    {
        state = src.state;
        buffer = src.buffer;
        //mSha = src.mSha;
        //memcpy(state.data(), src.state.data(), sizeof(uint32_t) * 5);
        //memcpy(block.data(), src.block.data(), sizeof(uint8_t) * 64);
        return *this;
    }
}