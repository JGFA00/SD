import poisson.*;

import java.lang.Thread;
import java.util.Random;

public class UsePoisson {

    static final int SAMPLES = 100;

    public static void main(String[] args) {
        if ( args.length != 1 )
            System.out.println("usage: poissonseq <lambda>");
        else
            testArrivalTimes(Double.parseDouble(args[0]));
    }

    private static void testArrivalTimes(double lambda) {
        PoissonProcess pp = new PoissonProcess(lambda, new Random(0) );
        for (int i = 1; i <= SAMPLES; i++) {
	    double t = pp.timeForNextEvent() * 60.0 * 1000.0;
            System.out.println("next event in -> " + (int)t + " ms");
	    try {
                Thread.sleep((int)t);
            }
            catch (InterruptedException e) {
                System.out.println("thread interrupted");
		e.printStackTrace(System.out);
            }
        }
    }
}

