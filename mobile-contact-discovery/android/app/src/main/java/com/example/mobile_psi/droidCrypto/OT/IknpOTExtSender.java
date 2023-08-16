package com.example.mobile_psi.droidCrypto.OT;

import com.example.mobile_psi.droidCrypto.Networking.Channel;
import java.nio.ByteBuffer;

public class IknpOTExtSender {

    static {
        System.loadLibrary("droidcrypto");
    }
    private long cNativeObj = 0;

    public IknpOTExtSender(ByteBuffer baseOts, byte[] baseChoices) {
        cNativeObj = init(baseOts, baseChoices);
    }


    public ByteBuffer send(int numOts,  Channel chan) {
        ByteBuffer output = ByteBuffer.allocateDirect(numOts * 2 * 128 / 8);
//        for(int i = 0; i < 16; i++) {
            send(cNativeObj, output, chan);
//        }
        return output;
    }

    public void cleanup() {
        deleteNativeObj(cNativeObj);
        cNativeObj = 0;
    }

    private native long init(ByteBuffer baseOTs, byte[] baseChoices);
    private native void send(long cNativeObj, ByteBuffer messages, Channel chan);
    private native void deleteNativeObj(long cNativeObj);
}
