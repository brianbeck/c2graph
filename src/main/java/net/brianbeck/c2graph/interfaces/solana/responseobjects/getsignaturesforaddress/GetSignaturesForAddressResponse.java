package net.brianbeck.c2graph.interfaces.solana.responseobjects.getsignaturesforaddress;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

public class GetSignaturesForAddressResponse {

    private String jsonrpc;

    @JsonProperty("result")
    private List<Signature> signatures;

    private int id;

    // Getters and Setters
    public String getJsonrpc() {
        return jsonrpc;
    }

    public void setJsonrpc(String jsonrpc) {
        this.jsonrpc = jsonrpc;
    }

    public List<Signature> getSignatures() {
        return signatures;
    }

    public void setSignatures(List<Signature> signatures) {
        this.signatures = signatures;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }
    
}
