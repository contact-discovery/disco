package com.example.mobile_psi.droidCrypto;

import android.os.AsyncTask;
import android.util.Log;

import com.contactdiscovery.mobile.MainActivity;

import client_android.Client_android;

public class TestPSIAsyncTask extends AsyncTask<Void, Void, Void> {

    private final String TAG = "NetworkTest";

    private int num_items;
    private int port;
    private long type;
    private String ip;
    private String response;

    static {
        System.loadLibrary("droidcrypto");
    }

    public TestPSIAsyncTask(int num_items, String ip, int port, long type) {
        this.num_items = num_items;
        this.ip = ip;
        this.type = type;
        this.port = port;
    }

    @Override
    protected Void doInBackground(Void... voids) {
        MainActivity.changeText("Running PSI for N_C="+ num_items );
        Log.v("PSI", num_items + " items.");
        response = TestAsyncTask.testNative(ip, port, type, num_items);
        return null;
    }

    @Override
    protected void onPostExecute(Void aVoid) {
        super.onPostExecute(aVoid);
        MainActivity.changeText("OPRF: " + response);
        Client_android.pirCallback(
                MainActivity.getPir1(),
                MainActivity.getPir2(),
                MainActivity.getClientExp(),
                MainActivity.getSegExp(),
                MainActivity.getDbExp(),
                prfTypeDroidToPIR(type),
                MainActivity.getMapping(),
                MainActivity.getWorkers(),
                false);
    }

    public static String prfTypeDroidToPIR(long prfDroid) {
        if (prfDroid == 2) {
            return "ECNR";
        } else if (prfDroid == 1) {
            return "GCAES";
        } else if (prfDroid == 0) {
            return "GCLOWMC";
        }
        return "";

    }

}
