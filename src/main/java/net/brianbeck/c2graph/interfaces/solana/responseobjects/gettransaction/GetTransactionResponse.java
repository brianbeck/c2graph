package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction;

import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;

public class GetTransactionResponse {

    private String jsonrpc;

    @JsonProperty("result")
    private GetTransactionResults transaction;

    private int id;

    // Getters and Setters
    public String getJsonrpc() {
        return jsonrpc;
    }

    public void setJsonrpc(String jsonrpc) {
        this.jsonrpc = jsonrpc;
    }

    public GetTransactionResults getTransaction() {
        return transaction;
    }

    public void setTransaction(GetTransactionResults transaction) {
        this.transaction = transaction;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public void printGetTransactionResponse() {

        System.out.println("ApiResponse:");
        System.out.println("  jsonrpc: " + this.getJsonrpc());
        System.out.println("  id: " + this.getId());

        GetTransactionResults result = this.getTransaction();
        if (result != null) {
            System.out.println("  Result:");
            System.out.println("  blockTime: " + result.getBlockTime());
            System.out.println("    slot: " + result.getSlot());

            // Meta
            Meta meta = result.getMeta();
            if (meta != null) {
                System.out.println("    Meta:");
                System.out.println("      err: " + meta.getErr());
                System.out.println("      fee: " + meta.getFee());
                System.out.println("      innerInstructions: " + meta.getInnerInstructions());
                System.out.println("      postBalances: " + meta.getPostBalances());
                System.out.println("      postTokenBalances: " + meta.getPostTokenBalances());
                System.out.println("      preBalances: " + meta.getPreBalances());
                System.out.println("      preTokenBalances: " + meta.getPreTokenBalances());
                System.out.println("      rewards: " + meta.getRewards());
                System.out.println("      compute units consumed: " + meta.getComputeUnitsConsumed());
                Status status = meta.getStatus();
                if (status != null) {
                    System.out.println("      Status:");
                    System.out.println("        Ok: " + status.getOk());
                }
            }

            // Transaction
            Transaction transaction = result.getTransaction();
            if (transaction != null) {
                System.out.println("    Transaction:");
                System.out.println("      Signatures: " + transaction.getSignatures());

                Message message = transaction.getMessage();
                if (message != null) {
                    System.out.println("      Message:");
                    System.out.println("        AccountKeys: " + message.getAccountKeys());
                    System.out.println("        RecentBlockhash: " + message.getRecentBlockhash());

                    Header header = message.getHeader();
                    if (header != null) {
                        System.out.println("        Header:");
                        System.out.println("          numReadonlySignedAccounts: " + header.getNumReadonlySignedAccounts());
                        System.out.println("          numReadonlyUnsignedAccounts: " + header.getNumReadonlyUnsignedAccounts());
                        System.out.println("          numRequiredSignatures: " + header.getNumRequiredSignatures());
                    }

                    List<Instruction> instructions = message.getInstructions();
                    if (instructions != null && !instructions.isEmpty()) {
                        System.out.println("        Instructions:");
                        for (Instruction instruction : instructions) {
                            System.out.println("          Accounts: " + instruction.getAccounts());
                            System.out.println("          Data: " + instruction.getData());
                            System.out.println("          ProgramIdIndex: " + instruction.getProgramIdIndex());
                        }
                    }
                }
            }
        } else {
            System.out.println("  Result: null");
        }
    }

}