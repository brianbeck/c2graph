package net.brianbeck.c2graph;

import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.Properties;

public class ApplicationProperties {
    private static volatile ApplicationProperties instance;

    private ApplicationProperties() {
        // private constructor to prevent instantiation
    }
    public static ApplicationProperties getInstance() {
        if (instance == null) {
            synchronized (ApplicationProperties.class) {
                if (instance == null) {
                    instance = new ApplicationProperties();

                    Properties environment = new Properties();

                    try (InputStream input = new FileInputStream("environment.properties")) {
                        environment.load(input);
                        // You can now access properties using prop.getProperty("key")
                    } catch (IOException ex) {
                        ex.printStackTrace();
                    }
                }
            }
        }
        return instance;
    }
}
