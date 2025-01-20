package net.brianbeck.c2graph;

    import java.io.FileWriter;
    import java.io.IOException;
    import java.sql.Connection;
    import java.sql.DriverManager;
    import java.sql.PreparedStatement;
    import java.sql.SQLException;
    import org.apache.commons.cli.*;

    import net.brianbeck.c2graph.interfaces.solana.SolanaInterface;
    import net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction.GetTransactionResponse;
    
    public class Main {
        public static void main(String[] args) {

            ApplicationProperties ap = ApplicationProperties.getInstance();

            String env = ap.getValue("ENVIRONMENT");
            //String env = ApplicationProperties.getInstance().getEnvironment().getProperty("ENVIRONMENT");

            if (env.equals("DEV")) {
                System.out.println("Application Started in DEV");

                SolanaInterface solanaInterface = new SolanaInterface();
        
                //solanaInterface.getAccountInformation("BLZZqav3Nxiza2pjjcD4hXGDqz8KKAKu3uDZpx39yQBt");
        
        
                //solanaInterface.getAccountInformation("HigCxjQYd93FxRM2BzQMTESfsVUQgHXjbp6YUzGkZrzw");
                //Signatures signatures = solanaInterface.getTransactions("HigCxjQYd93FxRM2BzQMTESfsVUQgHXjbp6YUzGkZrzw");
        
                GetTransactionResponse transactionDetails = solanaInterface.getTransactionDetails("25BYGLdqWwffKmd7HBPHx3Vy3QMbZXVZwEiuBh7WEK4QkdCVKuWKTiuaH2qkNtJk4H4LpjTE2SZyFuWKPzCQxTQ8");
                transactionDetails.printGetTransactionResponse();
                System.out.println("Application Completed");
                System.exit(0);

            } else if (env.equals("TEST")) {
                System.out.println("Application Started in TEST. Not currently supported.");
                System.out.println("Application Closing");
                System.exit(0);
            } else if (env.equals("PROD")) {
                System.out.println("Application Started in PROD. Not currently supported.");
                System.out.println("Application Closing");
                System.exit(0);
            } else {
                System.out.println("Application Started in UNKNOWN. Not currently supported.");
                System.out.println("Application Closing");
                System.exit(0);
            }

            Options options = new Options();
    
            Option outputFormat = new Option("f", "format", true, "output format: json or db");
            outputFormat.setRequired(true);
            options.addOption(outputFormat);
    
            Option outputDir = new Option("d", "directory", true, "output directory (required if format is json)");
            outputDir.setRequired(false);
            options.addOption(outputDir);
    
            Option dbUrl = new Option("u", "dburl", true, "database URL (required if format is db)");
            dbUrl.setRequired(false);
            options.addOption(dbUrl);
    
            Option dbUser = new Option("n", "dbuser", true, "database user (required if format is db)");
            dbUser.setRequired(false);
            options.addOption(dbUser);
    
            Option dbPassword = new Option("p", "dbpassword", true, "database password (required if format is db)");
            dbPassword.setRequired(false);
            options.addOption(dbPassword);
    
            CommandLineParser parser = new DefaultParser();
            HelpFormatter formatter = new HelpFormatter();
            CommandLine cmd;
    
            try {
                cmd = parser.parse(options, args);
            } catch (ParseException e) {
                System.out.println(e.getMessage());
                formatter.printHelp("c2graph", options);
                System.exit(1);
                return;
            }
    
            String format = cmd.getOptionValue("format");
            String directory = cmd.getOptionValue("directory");
            String dbUrlValue = cmd.getOptionValue("dburl");
            String dbUserValue = cmd.getOptionValue("dbuser");
            String dbPasswordValue = cmd.getOptionValue("dbpassword");
    
            SolanaInterface solanaInterface = new SolanaInterface();
            GetTransactionResponse transactionDetails = solanaInterface.getTransactionDetails("25BYGLdqWwffKmd7HBPHx3Vy3QMbZXVZwEiuBh7WEK4QkdCVKuWKTiuaH2qkNtJk4H4LpjTE2SZyFuWKPzCQxTQ8");
            transactionDetails.printGetTransactionResponse();
    
            if ("json".equalsIgnoreCase(format)) {
                if (directory == null) {
                    System.out.println("Output directory is required when format is json");
                    System.exit(1);
                }
                try (FileWriter file = new FileWriter(directory + "/transactionDetails.json")) {
                    //TODO: Convert Object to JSON
                    System.out.println("Successfully wrote JSON to file");
                } catch (IOException e) {
                    e.printStackTrace();
                }
            } else if ("db".equalsIgnoreCase(format)) {
                if (dbUrlValue == null || dbUserValue == null || dbPasswordValue == null) {
                    System.out.println("Database URL, user, and password are required when format is db");
                    System.exit(1);
                }
                try (Connection connection = DriverManager.getConnection(dbUrlValue, dbUserValue, dbPasswordValue)) {
                    String insertSQL = "INSERT INTO transactions (details) VALUES (?)";
                    try (PreparedStatement preparedStatement = connection.prepareStatement(insertSQL)) {
                        //TODO: convert object into format to insert in graph database
                        System.out.println("Successfully inserted transaction details into database");
                    }
                } catch (SQLException e) {
                    e.printStackTrace();
                }
            } else {
                System.out.println("Invalid format specified");
                System.exit(1);
            }
    
            transactionDetails.printGetTransactionResponse();
            System.out.println("Application Completed");
    }
}


