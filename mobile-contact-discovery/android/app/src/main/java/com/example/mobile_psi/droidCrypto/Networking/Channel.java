package com.example.mobile_psi.droidCrypto.Networking;

import android.annotation.TargetApi;

import java.io.IOException;
import java.net.SocketAddress;
import java.nio.ByteBuffer;
import java.nio.channels.AsynchronousServerSocketChannel;
import java.nio.channels.AsynchronousSocketChannel;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

public class Channel {

    public enum ROLE { SERVER, CLIENT }

    private AsynchronousSocketChannel socketChannel;
    private final String TAG = "Channel";
    private long sentBytes;
    private long recvBytes;

    @TargetApi(26)
    public Channel(SocketAddress addr, ROLE role) {
        if (role == ROLE.SERVER) {
            try {
                AsynchronousServerSocketChannel serverSocketChannel = AsynchronousServerSocketChannel.open().bind(addr);
                socketChannel = serverSocketChannel.accept().get();
                serverSocketChannel.close();
            } catch (IOException e) {
                e.printStackTrace();
            } catch (InterruptedException e) {
                e.printStackTrace();
            } catch (ExecutionException e) {
                e.printStackTrace();
            }
        } else if (role == ROLE.CLIENT) {
            try {
                socketChannel = AsynchronousSocketChannel.open();
                socketChannel.connect(addr).get();
            } catch (IOException e) {
                e.printStackTrace();
            } catch (InterruptedException e) {
                e.printStackTrace();
            } catch (ExecutionException e) {
                e.printStackTrace();
            }
        }
        resetCount();
    }

    @TargetApi(26)
    public Future<Integer> recvAsync(ByteBuffer buffer) {
        recvBytes += buffer.limit();
        return socketChannel.read(buffer);
    }

    @TargetApi(26)
    public void recv(ByteBuffer buffer) {
        try {
            recvBytes += buffer.limit();
            socketChannel.read(buffer).get();
            //Log.d(TAG, "recv:" + buffer.toString());
        } catch (InterruptedException e) {
            e.printStackTrace();
        } catch (ExecutionException e) {
            e.printStackTrace();
        }
    }

    @TargetApi(26)
    public Future<Integer> sendAsync(byte[] buffer) {
        ByteBuffer tmp = ByteBuffer.wrap(buffer);
        sentBytes += buffer.length;
        return socketChannel.write(tmp);
    }

    @TargetApi(26)
    public void sendAsyncVoid(byte[] buffer) {
        ByteBuffer tmp = ByteBuffer.wrap(buffer);
        sentBytes += buffer.length;
        socketChannel.write(tmp);
    }

    @TargetApi(26)
    public void send(ByteBuffer buffer) {
        try {
            sentBytes += buffer.limit();
            socketChannel.write(buffer).get();
            //Log.d(TAG, "sent" +  buffer.toString());
        } catch (InterruptedException e) {
            e.printStackTrace();
        } catch (ExecutionException e) {
            e.printStackTrace();
        }
    }

    public long getSentBytes() {
        return sentBytes;
    }

    public long getRecvBytes() {
        return recvBytes;
    }

    public void resetCount() {
        recvBytes = 0;
        sentBytes = 0;
    }
}
