package net.brianbeck.c2graph;

import net.brianbeck.c2graph.interfaces.solana.SolanaInterface;
import net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction.GetTransactionResponse;

public class Main {
    public static void main(String[] args) {
        System.out.println("Application Started");
        SolanaInterface solanaInterface = new SolanaInterface();

        //solanaInterface.getAccountInformation("BLZZqav3Nxiza2pjjcD4hXGDqz8KKAKu3uDZpx39yQBt");


        //solanaInterface.getAccountInformation("HigCxjQYd93FxRM2BzQMTESfsVUQgHXjbp6YUzGkZrzw");
        //Signatures signatures = solanaInterface.getTransactions("HigCxjQYd93FxRM2BzQMTESfsVUQgHXjbp6YUzGkZrzw");

        GetTransactionResponse transactionDetails = solanaInterface.getTransactionDetails("25BYGLdqWwffKmd7HBPHx3Vy3QMbZXVZwEiuBh7WEK4QkdCVKuWKTiuaH2qkNtJk4H4LpjTE2SZyFuWKPzCQxTQ8");
        transactionDetails.printGetTransactionResponse();
        System.out.println("Application Completed");
    }
}