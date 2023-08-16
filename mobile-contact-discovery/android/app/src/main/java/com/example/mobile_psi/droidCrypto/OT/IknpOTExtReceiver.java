package com.example.mobile_psi.droidCrypto.OT;

import com.example.mobile_psi.droidCrypto.Networking.Channel;

import java.nio.ByteBuffer;

public class IknpOTExtReceiver {

    static {
        System.loadLibrary("droidcrypto");
    }

    private long cNativeObj = 0;

    public IknpOTExtReceiver(ByteBuffer baseOts) {
        cNativeObj = init(baseOts);
    }


    public ByteBuffer recv(byte[] choices, Channel chan) {
        ByteBuffer output = ByteBuffer.allocateDirect(choices.length*8*128/8);
//        for(int i = 0; i < 16; i++) {
            recv(cNativeObj, output, choices, chan);
//        }
        return output;

    }

    public void cleanup() {
        deleteNativeObj(cNativeObj);
        cNativeObj = 0;
    }

    private native long init(ByteBuffer baseOTs);
    private native void recv(long cNativeObj, ByteBuffer messages, byte[] choices, Channel chan);
    private native void deleteNativeObj(long cNativeObj);

}
