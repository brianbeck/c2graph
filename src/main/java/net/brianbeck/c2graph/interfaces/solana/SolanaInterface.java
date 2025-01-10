package net.brianbeck.c2graph.interfaces.solana;

import java.util.Properties;

import javax.ws.rs.client.Client;
import javax.ws.rs.client.ClientBuilder;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;
import javax.ws.rs.client.Entity;

import net.brianbeck.c2graph.interfaces.CryptoInterface;
import net.brianbeck.c2graph.interfaces.solana.responseobjects.getaccountinfo.GetAccountInformationResponse;
import net.brianbeck.c2graph.interfaces.solana.responseobjects.getsignaturesforaddress.GetSignaturesForAddressResponse;
import net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction.GetTransactionResponse;

public class SolanaInterface implements CryptoInterface {

    Properties solanaProps = new Properties();

    public SolanaInterface() {

/*         try (InputStream input = new FileInputStream("solana.properties")) {
            solanaProps.load(input);
            // You can now access properties using prop.getProperty("key")
        } catch (IOException ex) {
            ex.printStackTrace();
        } */
    }

    public GetAccountInformationResponse getAccountInformation(String accountId) {

        GetAccountInformationResponse deserializedResponse = null;

        try {
            Client client = ClientBuilder.newClient();

            // Define the target URL
            String apiUrl = "https://api.devnet.solana.com";
            //String apiUrl = "https://api.mainnet-beta.solana.com";

            // Define the payload (JSON)
            String jsonPayload = "{"
                + "\"jsonrpc\": \"2.0\","
                + "\"id\": 1,"
                + "\"method\": \"getAccountInfo\","
                + "\"params\": ["
                + "\""
                + accountId
                + "\", { "
                + "\"encoding\": \"base58\""
                + "} ]"
                + "}";

            System.out.println("Payload: " + jsonPayload);

            // Make a POST request with the payload
            Response response = client.target(apiUrl)
                    .request(MediaType.APPLICATION_JSON)
                    .post(Entity.entity(jsonPayload, MediaType.APPLICATION_JSON));

            // Check status and read the response
            if (response.getStatus() == 200) {
                //String responseBody = response.readEntity(String.class);
                //System.out.println("Response: " + responseBody);

                deserializedResponse = response.readEntity(GetAccountInformationResponse.class);

                // Print the deserialized object
                System.out.println("Deserialized Response:");
                System.out.println("JSON-RPC: " + deserializedResponse.getJsonrpc());
                System.out.println("ID: " + deserializedResponse.getId());
                System.out.println("Lamports: " + deserializedResponse.getResult().getValue().getLamports());
                System.out.println("Rent Epoch: " + deserializedResponse.getResult().getValue().getRentEpoch());

            } else {
                System.out.println("Error: HTTP " + response.getStatus());
                System.out.println("Response: " + response.readEntity(String.class));
            }

        // Close the client
        client.close();

        } catch (Exception e) {
            e.printStackTrace();
        }

        return deserializedResponse;

    }

    public GetSignaturesForAddressResponse getTransactions(String accountId) {

        GetSignaturesForAddressResponse deserializedResponse = null;

        try {
            Client client = ClientBuilder.newClient();

            // Define the target URL
            String apiUrl = "https://api.devnet.solana.com";
            //String apiUrl = "https://api.mainnet-beta.solana.com";

            // TODO: Add support for before and until
            // Define the payload (JSON)
            String jsonPayload = "{"
                + "\"jsonrpc\": \"2.0\","
                + "\"id\": 1,"
                + "\"method\": \"getSignaturesForAddress\","
                + "\"params\": ["
                + "\""
                + accountId
                + "\","
                + "{"
                + "\"limit\": 1000"
                + "}"
                + "]"
                + "}";

            System.out.println("Payload: " + jsonPayload);

            // Make a POST request with the payload
            Response response = client.target(apiUrl)
                    .request(MediaType.APPLICATION_JSON)
                    .post(Entity.entity(jsonPayload, MediaType.APPLICATION_JSON));

            // Print the response status
            System.out.println("HTTP Status: " + response.getStatus());

            // Read and print the JSON response
            //String jsonResponse = response.readEntity(String.class);
            //System.out.println("JSON Response: ");
            //System.out.println(jsonResponse);

            // Check status and read the response
            if (response.getStatus() == 200) {

                deserializedResponse = response.readEntity(GetSignaturesForAddressResponse.class);

                // Print the deserialized object
                System.out.println("Deserialized Response:");
                System.out.println("JSON-RPC: " + deserializedResponse.getJsonrpc());
                System.out.println("ID: " + deserializedResponse.getId());
                System.out.println("Number of Signatures: " + deserializedResponse.getSignatures().toArray().length);

                //for (int i = 0; i < deserializedResponse.getSignatures().toArray().length; i++) {
                //    System.out.println("Signature " + i + ": " + deserializedResponse.getSignatures().get(i).getSignature());
                //    System.out.println("Signature " + i + ": " + deserializedResponse.getSignatures().get(i).getConfirmationStatus());
                //    System.out.println("Signature " + i + ": " + deserializedResponse.getSignatures().get(i).getSlot());
                //}

            } else {
                System.out.println("Error: HTTP " + response.getStatus());
                System.out.println("Response: " + response.readEntity(String.class));
            }

        // Close the client
        client.close();

        } catch (Exception e) {
            e.printStackTrace();
        }

        return deserializedResponse;

    }

    public GetTransactionResponse getTransactionDetails(String transactionID) {

        GetTransactionResponse deserializedResponse = null;

        try {
            Client client = ClientBuilder.newClient();

            // Define the target URL
            String apiUrl = "https://api.devnet.solana.com";
            //String apiUrl = "https://api.mainnet-beta.solana.com";

            // TODO: Add support for before and until
            // Define the payload (JSON)
            String jsonPayload = "{"
                + "\"jsonrpc\": \"2.0\","
                + "\"id\": 1,"
                + "\"method\": \"getTransaction\","
                + "\"params\": ["
                + "\""
                + transactionID 
                + "\","
                + "\"json\""
                + "]"
                + "}";

            System.out.println("Payload: " + jsonPayload);

            // Make a POST request with the payload
            Response response = client.target(apiUrl)
                    .request(MediaType.APPLICATION_JSON)
                    .post(Entity.entity(jsonPayload, MediaType.APPLICATION_JSON));

            // Print the response status
            System.out.println("HTTP Status: " + response.getStatus());

            // Read and print the JSON response
            //String jsonResponse = response.readEntity(String.class);
            //System.out.println("JSON Response: ");
            //System.out.println(jsonResponse);

            // Check status and read the response
            if (response.getStatus() == 200) {

                deserializedResponse = response.readEntity(GetTransactionResponse.class);

                // Print the deserialized object
                System.out.println("Deserialized Response:");


            } else {
                System.out.println("Error: HTTP " + response.getStatus());
                System.out.println("Response: " + response.readEntity(String.class));
            }

        // Close the client
        client.close();

        } catch (Exception e) {
            e.printStackTrace();
        }

        return deserializedResponse;

    }

}
