package net.brianbeck.c2graph.interfaces.solana.responseobjects.getaccountinfo;

public class Context {
    private String apiVersion;
    private long slot;

    // Getters and Setters
    public String getApiVersion() {
        return apiVersion;
    }

    public void setApiVersion(String apiVersion) {
        this.apiVersion = apiVersion;
    }

    public long getSlot() {
        return slot;
    }

    public void setSlot(long slot) {
        this.slot = slot;
    }
}
