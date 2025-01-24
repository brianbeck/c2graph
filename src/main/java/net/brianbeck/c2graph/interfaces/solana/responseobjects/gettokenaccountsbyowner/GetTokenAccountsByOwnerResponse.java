package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettokenaccountsbyowner;

import com.fasterxml.jackson.annotation.JsonProperty;

public class GetTokenAccountsByOwnerResponse {

    private String jsonrpc;

    @JsonProperty("result")
    private GetTokenAcountsByOwnerResult result;
    private int id;

    // Getters and Setters
    public String getJsonrpc() {
        return jsonrpc;
    }

    public void setJsonrpc(String jsonrpc) {
        this.jsonrpc = jsonrpc;
    }

    public GetTokenAcountsByOwnerResult getResult() {
        return result;
    }

    public void setResult(GetTokenAcountsByOwnerResult result) {
        this.result = result;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }
}
