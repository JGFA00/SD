package ds.examples.sockets.calculator;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.Scanner;

public class Server {
    private ServerSocket server;

    public Server(String ipAddress) throws Exception {
        this.server = new ServerSocket(0, 1, InetAddress.getByName(ipAddress));
    }

    private void listen() throws Exception {
	String command;
	double result;
	Socket client = this.server.accept();
	String clientAddress = client.getInetAddress().getHostAddress();
	System.out.printf("\r\nnew connection from %s\n", clientAddress);
            
	/*
         * prepare socket I/O channels
         */
	BufferedReader in = new BufferedReader(new InputStreamReader(client.getInputStream()));        
	PrintWriter out = new PrintWriter(client.getOutputStream(), true);

	while(true) {
	    /* 
	     * receive command 
	     */
	    if( (command = in.readLine()) == null)
		break;
	    else
		System.out.printf("message from %s : %s\n", clientAddress, command);

	    /*
	     * process command
	     */
	    result = process(command);

	    /*
	     * send result
	     */
	    out.println(String.valueOf(result));
	    out.flush();
	}
    }

    private double process(String command) {
	    Scanner sc = new Scanner(command).useDelimiter(":");
	    String  op = sc.next();
	    double  x  = Double.parseDouble(sc.next());
	    double  y  = Double.parseDouble(sc.next());
	    double  z  = 0.0; 
	    switch(op) {
	    case "add": z = x + y; break;
	    case "sub": z = x - y; break;
	    case "mul":	z = x * y; break;
	    case "div":	z = x / y; break;
	    }
	    return z;
    }
    
    public InetAddress getSocketAddress() {
	return this.server.getInetAddress();
    }
        
    public int getPort() {
	return this.server.getLocalPort();
    }
    
    public static void main(String[] args) throws Exception {
	Server app = new Server(args[0]);
	System.out.printf("\r\nrunning server: host=%s @ port=%d\n",
	    app.getSocketAddress().getHostAddress(), app.getPort());
	app.listen();
    }
}
