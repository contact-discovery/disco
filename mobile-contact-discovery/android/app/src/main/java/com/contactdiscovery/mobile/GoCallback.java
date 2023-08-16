package com.contactdiscovery.mobile;


import android.util.Log;
import client_android.Client_android;
import client_android.JavaCallback;

public class GoCallback implements JavaCallback {

    public void callFromGo(String data) {
        Log.d("[GO CALLBACK]", data);
        MainActivity.changeText(data);
    }

}
