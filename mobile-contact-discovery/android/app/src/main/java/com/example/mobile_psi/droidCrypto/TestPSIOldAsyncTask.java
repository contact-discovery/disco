package com.example.mobile_psi.droidCrypto;

import android.os.AsyncTask;
import android.util.Log;

import com.contactdiscovery.mobile.MainActivity;

public class TestPSIOldAsyncTask extends AsyncTask<Void, Void, Void> {

    private final String TAG = "NetworkTest";

    private int num_items;
    private int port;
    private long type;
    private String ip;
    private String response;

    static {
        System.loadLibrary("droidcrypto");
    }

    public TestPSIOldAsyncTask (int num_items, String ip, int port, long type) {
        this.num_items = num_items;
        this.ip = ip;
        this.type = type;
        this.port = port;
    }

    @Override
    protected Void doInBackground(Void... voids) {
        MainActivity.changeText("Running [KRS+19] PSI for N_C="+ num_items );
        Log.v("PSI", num_items + " items.");
        response = testNative(ip, port, type, num_items);
        return null;
    }

    @Override
    protected void onPostExecute(Void aVoid) {
        super.onPostExecute(aVoid);
        MainActivity.changeText("PSI [KRS+19]: "+ response);
    }

    private native String testNative(String ip, int port, long type, int num_items);
}
