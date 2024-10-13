/*
 * adapted from https://www.pegaxchange.com/2017/12/07/simple-tcp-ip-server-client-java/
 */

package ds.examples.sockets.basic;

import java.io.IOException;
import java.io.PrintWriter;
import java.net.InetAddress;
import java.net.Socket;
import java.util.Scanner;

public class Client {
    private Socket socket;
    private Scanner scanner;
    
    private Client(InetAddress serverAddress, int serverPort) throws Exception {
	this.socket = new Socket(serverAddress, serverPort);
	this.scanner = new Scanner(System.in);
    }

    private void start() throws IOException {
	int i = 0;
	String input;
	while (true) {
	    System.out.printf("[%3d]$ ",i);
	    input = scanner.nextLine();
	    if(input.equals("quit"))
		break;
	    PrintWriter out = new PrintWriter(this.socket.getOutputStream(), true);
	    out.println(input);
	    out.flush();
	    i++;
	}
    }
        
    public static void main(String[] args) throws Exception {
	Client client = new Client(InetAddress.getByName(args[0]), Integer.parseInt(args[1]));
	System.out.printf("\r\nconnected to server: %s\n", client.socket.getInetAddress());
	client.start();                
    }
}
