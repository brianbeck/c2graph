package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction;

import java.util.List;

public class Transaction {
    private Message message;
    private List<String> signatures;

    // Getters and Setters
    public Message getMessage() {
        return message;
    }

    public void setMessage(Message message) {
        this.message = message;
    }

    public List<String> getSignatures() {
        return signatures;
    }

    public void setSignatures(List<String> signatures) {
        this.signatures = signatures;
    }
}