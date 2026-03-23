package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction;

import com.fasterxml.jackson.annotation.JsonProperty;

public class Status {

    @JsonProperty("Ok")
    private Object Ok;

    // Getters and Setters
    public Object getOk() {
        return Ok;
    }

    public void setOk(Object Ok) {
        this.Ok = Ok;
    }
}
