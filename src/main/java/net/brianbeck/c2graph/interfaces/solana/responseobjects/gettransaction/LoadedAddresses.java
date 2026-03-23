package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction;

import java.util.List;

public class LoadedAddresses {
    private List<String> writable;
    private List<String> readonly;

    // Getters and Setters
    public List<String> getWritable() {
        return writable;
    }

    public void setWritable(List<String> writable) {
        this.writable = writable;
    }

    public List<String> getReadonly() {
        return readonly;
    }

    public void setReadonly(List<String> readonly) {
        this.readonly = readonly;
    }
}