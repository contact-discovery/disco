package com.example.mobile_psi.droidCrypto.OT;

import com.example.mobile_psi.droidCrypto.Networking.Channel;

import java.nio.ByteBuffer;

public class NaorPinkas {

    static {
        System.loadLibrary("droidcrypto");
    }

    public NaorPinkas() {}

    public ByteBuffer recv(byte[] choices, Channel chan) {
        ByteBuffer output = ByteBuffer.allocateDirect(choices.length*8 * 128 / 8);
        recv(output, choices, chan);
        return output;
    }

    public ByteBuffer send(int numOts,  Channel chan) {
        ByteBuffer output = ByteBuffer.allocateDirect(numOts * 2 * 128 / 8);
        send(output, chan);
        return output;
    }

    private native void send(ByteBuffer messages, Channel chan);
    private native void recv(ByteBuffer messages, byte[] choices, Channel chan);
}
