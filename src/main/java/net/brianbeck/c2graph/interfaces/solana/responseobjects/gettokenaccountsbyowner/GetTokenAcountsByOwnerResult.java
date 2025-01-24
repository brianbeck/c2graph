package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettokenaccountsbyowner;

import java.util.List;

public class GetTokenAcountsByOwnerResult {

    private Context context;
    private List<Value> value;

    public Context getContext() {
        return context;
    }

    public void setContext(Context context) {
        this.context = context;
    }

    public List<Value> getValue() {
        return value;
    }

    public void setValue(List<Value> value) {
        this.value = value;
    }
}