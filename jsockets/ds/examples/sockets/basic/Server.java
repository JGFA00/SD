/*
* adapted from https://www.pegaxchange.com/2017/12/07/simple-tcp-ip-server-client-java/
*/

package ds.examples.sockets.basic;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;

public class Server {
  private ServerSocket server;

  public Server(String ipAddress) throws Exception {
    this.server = new ServerSocket(0, 1, InetAddress.getByName(ipAddress));
  }

  private void listen() throws Exception {
    String data = null;
    Socket client = this.server.accept();
    String clientAddress = client.getInetAddress().getHostAddress();
    System.out.printf("\r\nnew connection from %s\n", clientAddress);

    BufferedReader in = new BufferedReader(new InputStreamReader(client.getInputStream()));
    while ( (data = in.readLine()) != null ) {
      System.out.printf("\r\nmessage from %s : %s\n", clientAddress, data);
    }
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
