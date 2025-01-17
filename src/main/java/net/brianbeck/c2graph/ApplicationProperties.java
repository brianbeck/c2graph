package net.brianbeck.c2graph;

import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.Properties;

public class ApplicationProperties {

    //private static final Logger LOG = LoggerFactory.getLogger(ApplicationProperties.class);

    private static ApplicationProperties _instance = null;
    private Properties _environmentProperties = null;

    private String environment = "ENVIRONMENT";
    private String dev = "DEV";
    private String test = "TEST";
    private String production = "PROD";


    protected ApplicationProperties() {
        _environmentProperties = new Properties();

        try (InputStream input = new FileInputStream("c2graph/src/main/resources/environment.properties")) {
            _environmentProperties.load(input);
            System.out.println("******* " + _environmentProperties.getProperty(environment));
        } catch (IOException ex) {
            ex.printStackTrace();
        }

        if (_environmentProperties.getProperty(environment).equals(dev)) {
            try (OutputStream input = new FileOutputStream("c2graph/src/main/resources/environment.properties")) {
                _environmentProperties.store(input, null);
            } catch (IOException ex) {
                ex.printStackTrace();
            }
        } else if (_environmentProperties.getProperty(environment).equals(production)) {
            try (OutputStream input = new FileOutputStream("c2graph/src/main/resources/environment.properties")) {
                _environmentProperties.store(input, null);
            } catch (IOException ex) {
                ex.printStackTrace();
            }
        } else if (_environmentProperties.getProperty(environment).equals(test)) {
            try (OutputStream input = new FileOutputStream("c2graph/src/main/resources/environment.properties")) {
                _environmentProperties.store(input, null);
            } catch (IOException ex) {
                ex.printStackTrace();
            }
        } else {
            //TODO: Log this and exit the application
            System.out.println("Unknown Environment");
        }

    }
    public static synchronized ApplicationProperties getInstance() {
        
        if (_instance == null) {
            synchronized (ApplicationProperties.class) {
                if (_instance == null) {
                    _instance = new ApplicationProperties();
                }
            }
        }
        return _instance;
    }

    public String getValue(String key) {
        return this._environmentProperties.getProperty(key, String.format("The key %s does not exists!", key));
    }
}
