package com.contactdiscovery.mobile;

import android.app.Activity;
import android.os.Bundle;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.Spinner;
import android.widget.TextView;


import com.example.mobile_psi.droidCrypto.TestAsyncTask;
import com.example.mobile_psi.droidCrypto.TestPSIAsyncTask;
import com.example.mobile_psi.droidCrypto.TestPSIOldAsyncTask;
import com.example.mobile_psi.droidCrypto.TestSpeedTask;

import client_android.Client_android;

public class MainActivity extends Activity {


    // Used to load the 'native-lib' library on application startup.
    static {
        System.loadLibrary("droidcrypto");
    }
    public Button buttonOPRF;
    Button buttonPSI;
    Button buttonPSI_old;
    Button buttonPartitionTest;
    Button buttonPIR;
    Button buttonSpeed;
    EditText pir1Text;
    EditText pir2Text;
    EditText oprfText;
    Spinner dbExpSpinner;
    Spinner segExpSpinner;
    Spinner clientExpSpinner;
    Spinner prfSpinner;
    EditText mappingText;
    EditText workersText;
    EditText onlinePhasesText;
    static TextView outputText;
    static GoCallback gocb;


    private static String oprf;
    private static String pir1;
    private static String pir2;
    private static String prfType;
    private static int clientExp;
    private static int segExp;
    private static int dbExp;
    private static double mapping;
    private static int workers;
    //private static int onlinePhases;


    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        buttonOPRF = findViewById(R.id.doOPRF);
        buttonPIR = findViewById(R.id.doPIR);
        buttonPSI = findViewById(R.id.doPSI);
        buttonPSI_old = findViewById(R.id.doPSI_old);
        buttonPartitionTest = findViewById(R.id.doPartitionTest);
        buttonSpeed= findViewById(R.id.doSpeedtest);
        pir1Text = findViewById(R.id.pir1);
        pir2Text = findViewById(R.id.pir2);
        oprfText = findViewById(R.id.oprf);
        dbExpSpinner = (Spinner)findViewById(R.id.dbExpSpinner);
        segExpSpinner = (Spinner)findViewById(R.id.segExpSpinner);
        clientExpSpinner = (Spinner)findViewById(R.id.clientExpSpinner);
        prfSpinner = (Spinner)findViewById(R.id.prfSpinner);
        mappingText = findViewById(R.id.mappingPercent);
        workersText = findViewById(R.id.numWorkers);
        outputText = findViewById(R.id.textOutput);
        //onlinePhasesText = findViewById(R.id.numOnlineRuns);
        System.out.println(getPir1());
        System.out.println(getOprf());
        gocb = new GoCallback();
        Client_android.registerJavaCallback(gocb);

        buttonPIR.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                SetParams();
                outputText.setText("Running extended Offline-Online PIR protocol...");
                Client_android.pirCallback(
                        getPir1(),
                        getPir2(),
                        getClientExp(),
                        getSegExp(),
                        getDbExp(),
                        PRFType(prfSpinner.getSelectedItem().toString()),
                        getMapping(),
                        getWorkers(),
                        false);
            }
        });

        buttonOPRF.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                SetParams();
                String[] oprfVals = getOprf().split(":");
                TestAsyncTask task = new TestAsyncTask(
                        getClientJava(),
                        oprfVals[0],
                        Integer.parseInt((oprfVals[1])),
                        PRFTypeDroidCrypto(getPrfType()));
                task.execute();
            }
        });

        buttonPartitionTest.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                SetParams();
                outputText.setText("Running Test with one DB Partition per Server...");
                Client_android.pirCallback(
                        getPir1(),
                        getPir2(),
                        getClientExp(),
                        getSegExp(),
                        getDbExp(),
                        PRFType(prfSpinner.getSelectedItem().toString()),
                        getMapping(),
                        getWorkers(),
                        true);
            }
        });

        buttonPSI.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                SetParams();
                String[] oprfVals = getOprf().split(":");
                TestPSIAsyncTask task = new TestPSIAsyncTask(
                        getClientExp(),
                        oprfVals[0],
                        Integer.parseInt((oprfVals[1])),
                        PRFTypeDroidCrypto(getPrfType()));
                task.execute();
            }
        });

        buttonPSI_old.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                SetParams();
                String[] oprfVals = oprf.split(":");
                TestPSIOldAsyncTask task = new TestPSIOldAsyncTask(
                        getClientJava(),
                        oprfVals[0],
                        Integer.parseInt((oprfVals[1])),
                        PRFTypeDroidCrypto(getPrfType()));
                task.execute();
            }
        });

        /*
        buttonPartitionTest.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                changeText("Not implemented yet");
            }
        });
*/

        buttonSpeed.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                changeText("Not implemented yet");
                /*
                SetParams();
                String[] oprfVals = oprf.split(":");
                TestSpeedTask task = new TestSpeedTask(
                        oprfVals[0],
                        Integer.parseInt((oprfVals[1]))
                );
                task.execute();
                */
            }
        });

    }
    public static void changeText(String text) {
        outputText.setText(text);
    }

    public static long PRFTypeDroidCrypto(String inType) {
        if ("EC-NR".equals(inType)) {
            return 2;
        } else if ("GC-AES".equals(inType)) {
            return 1;
        } else if ("GC-LowMC".equals(inType)) {
            return 0;
        }
        return 99;
    }

    public static String PRFType(String inType) {
        if ("EC-NR".equals(inType)) {
            return "ECNR";
        } else if ("GC-AES".equals(inType)) {
            return "GCAES";
        } else if ("GC-LowMC".equals(inType)) {
            return "GCLOWMC";
        }
        return "";
    }

    public void SetParams() {

        pir1 = pir1Text.getText().toString();
        if ("".equals(pir1)) {
            pir1 = "130.83.125.183:50052";
        }
        pir2 = pir2Text.getText().toString();
        if ("".equals(pir2)) {
            pir2 = "130.83.125.184:50052";
        }
        oprf = oprfText.getText().toString();
        if ("".equals(oprf)) {
            oprf = "130.83.125.183:50051";
        }
        mapping = Float.parseFloat(mappingText.getText().toString());
        if ("".equals(mapping)) {
            mapping = 0.99;
        }
        workers = Integer.parseInt(workersText.getText().toString());
        if ("".equals(workers)) {
            workers = 8;
        }
        /*
        onlinePhases = Integer.parseInt(onlinePhasesText.getText().toString());
        if ("".equals(onlinePhases)) {
            onlinePhases = 1;
        }
        */
        prfType = prfSpinner.getSelectedItem().toString();
        clientExp = Integer.parseInt(clientExpSpinner.getSelectedItem().toString());
        segExp = Integer.parseInt(segExpSpinner.getSelectedItem().toString());
        dbExp = Integer.parseInt(dbExpSpinner.getSelectedItem().toString());
    }

    public String getOprf() {
        return oprf;
    }

    public static String getPir1() {
        return pir1;
    }

    public static String getPir2() {
        return pir2;
    }

    public static String getPrfType() {
        return prfType;
    }

    public static int getClientJava() {
        if (clientExp == 0) {
            return 1;
        } else if (clientExp == 10) {
            return 1024;
        }else if (clientExp == 14) {
            return 16384;
        }
        return 0;
    }

    public static int getClientExp() {
        return clientExp;
    }

    public static int getSegExp() {
        return segExp;
    }

    public static int getDbExp() {
        return dbExp;
    }

    public static double getMapping() {
        return mapping;
    }

    public static int getWorkers() {
        return workers;
    }

}
