package com.example.mobile_psi.droidCrypto;

import android.os.AsyncTask;

import com.contactdiscovery.mobile.MainActivity;

public class TestSpeedTask extends AsyncTask<Void, Void, Void> {

    private int port;
    private String ip;
    private String response;

    static {
        System.loadLibrary("droidcrypto");
    }

    public TestSpeedTask(String ip, int port) {
        this.ip = ip;
        this.port = port;
    }

    @Override
    protected Void doInBackground(Void... voids) {
        MainActivity.changeText("Running Speedtest");
        testSpeed(ip, port);
        return null;
    }

    @Override
    protected void onPostExecute(Void aVoid) {
        super.onPostExecute(aVoid);
    }



    public native void testSpeed(String ip, int port);
}
