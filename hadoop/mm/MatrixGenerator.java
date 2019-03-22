
import java.io.File;
import java.io.IOException;
import java.text.Format;
import java.time.LocalDateTime;
import java.util.Random;
import org.apache.commons.io.FileUtils;

/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */

/**
 *
 * @author rcarlini
 */
public class MatrixGenerator {
    
    private static int MAX_RANDOM = 9;
    private static Random rand = new Random(System.currentTimeMillis());
    
    public static void main(String[] args) throws IOException {
        if (args.length < 4) {
            System.out.println("Not enough arguments!");
        } else {
            
            int aRows = Integer.parseInt(args[0]);
            int aColumns = Integer.parseInt(args[1]);
            int bRows = Integer.parseInt(args[2]);
            int bColumns = Integer.parseInt(args[3]);

            String filename;
            if (args.length > 4) {
                filename = args[4];
            } else {
                filename = "" + aRows + "_" + aColumns + "_" + bColumns + ".txt";
            }
            File outFile = new File(filename);            
            FileUtils.writeStringToFile(outFile, "");

            generateMatrix("A", aRows, aColumns, outFile);
            generateMatrix("B", bRows, bColumns, outFile);
        }
    }

    
    private static void generateMatrix(String id, int rows, int columns, File outFile) throws IOException {
        for(int i = 0; i<rows; i++) {
            for (int j = 0; j < columns; j++) {
                int value = rand.nextInt(MAX_RANDOM);
                String entry = id + " " + rows + " " + columns + " " + i + " " + j + " " + value + "\n";
                FileUtils.writeStringToFile(outFile, entry, true);
            }
        }
    }
}
