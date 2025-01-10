package net.brianbeck.c2graph.interfaces.solana.responseobjects.getaccountinfo;

public class GetAccountInformationResponse {

    private String jsonrpc;
    private GetAccountInformationResult result;
    private int id;

    // Getters and Setters
    public String getJsonrpc() {
        return jsonrpc;
    }

    public void setJsonrpc(String jsonrpc) {
        this.jsonrpc = jsonrpc;
    }

    public GetAccountInformationResult getResult() {
        return result;
    }

    public void setResult(GetAccountInformationResult result) {
        this.result = result;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }
}
