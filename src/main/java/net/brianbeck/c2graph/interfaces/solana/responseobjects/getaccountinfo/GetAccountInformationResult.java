package net.brianbeck.c2graph.interfaces.solana.responseobjects.getaccountinfo;

public class GetAccountInformationResult {

    private Context context;
    private Value value;

    // Getters and Setters
    public Context getContext() {
        return context;
    }

    public void setContext(Context context) {
        this.context = context;
    }

    public Value getValue() {
        return value;
    }

    public void setValue(Value value) {
        this.value = value;
    }
}