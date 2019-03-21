import java.io.DataInput;
import java.io.DataOutput;
import java.io.IOException;
import java.util.ArrayList;
import java.util.StringTokenizer;

import org.apache.hadoop.conf.Configuration;
import org.apache.hadoop.fs.Path;
import org.apache.hadoop.io.IntWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.io.Writable;
import org.apache.hadoop.io.WritableUtils;
import org.apache.hadoop.mapreduce.Job;
import org.apache.hadoop.mapreduce.Mapper;
import org.apache.hadoop.mapreduce.Reducer;
import org.apache.hadoop.mapreduce.lib.input.TextInputFormat;
import org.apache.hadoop.mapreduce.lib.output.FileOutputFormat;
import org.apache.hadoop.mapreduce.lib.reduce.IntSumReducer;

/**
 *
 * @author rcarlini
 */
public class MatrixMultiplier {

    public static class CellWritable implements Writable {

        private char matrix;
        private int rows;
        private int columns;
        private int row;
        private int column;
        private int value;


        public CellWritable(CellWritable other) {
            this.matrix = other.matrix;
            this.rows = other.rows;
            this.columns = other.columns;
            this.row = other.row;
            this.column = other.column;
            this.value = other.value;
        }

        public CellWritable() {
        }

        private void set(String[] pieces) {
            this.matrix  = pieces[0].charAt(0);
            this.rows    = Integer.parseInt(pieces[1]);
            this.columns = Integer.parseInt(pieces[2]);
            this.row     = Integer.parseInt(pieces[3]);
            this.column  = Integer.parseInt(pieces[4]);
            this.value   = Integer.parseInt(pieces[5]);
        }

        @Override
        public void write(DataOutput d) throws IOException {
            d.writeChar(this.matrix);
            d.writeInt(this.rows);
            d.writeInt(this.columns);
            d.writeInt(this.row);
            d.writeInt(this.column);
            d.writeInt(this.value);
        }

        @Override
        public void readFields(DataInput di) throws IOException {
            this.matrix = di.readChar();
            this.rows = di.readInt();
            this.columns = di.readInt();
            this.row = di.readInt();
            this.column = di.readInt();
            this.value = di.readInt();
        }
    }

    public static class MatrixMapper
            extends Mapper<Object, Text, Text, CellWritable> {

        private CellWritable cell = new CellWritable();
        private Text index = new Text();
        
        public void map(Object key, Text value, Context context) throws IOException, InterruptedException {
            
            String[] pieces = value.toString().split(" ");
            cell.set(pieces);
           
            if ('A' == cell.matrix) {
                index.set("C " + cell.column);
                context.write(index, cell);
            } else {
                index.set("C " + cell.row);
                context.write(index, cell);
            }
        }
    }


    public static class MyIdentityMapper
            extends Mapper<Object, Text, Text, IntWritable> {

        private Text index = new Text();
        private IntWritable wValue = new IntWritable();
        
        public void map(Object key, Text value, Context context) throws IOException, InterruptedException {
            
            String[] pieces = value.toString().split("\t");
            
            index.set(pieces[0]);
            wValue.set(Integer.parseInt(pieces[1]));
            
            context.write(index, wValue);
        }
    }

    public static class MultiplierReducer
            extends Reducer<Text, CellWritable, Text, IntWritable> {

        private Text index = new Text();
        private IntWritable result = new IntWritable();

        public void reduce(Text key, Iterable<CellWritable> values, Context context) throws IOException, InterruptedException {
            
            ArrayList<CellWritable> aCells = new ArrayList<>();
            ArrayList<CellWritable> bCells = new ArrayList<>();

            for (CellWritable cell : values) {
                
                CellWritable newCell = new CellWritable(cell);
                if ('A' == cell.matrix) {
                    aCells.add(newCell);
                } else {
                    bCells.add(newCell);
                }
            }
            
            for (CellWritable aCell : aCells) {
                for (CellWritable bCell : bCells) {
                    String cKey = "C " + aCell.row + " " + bCell.column;

                    index.set(cKey);
                    result.set(aCell.value * bCell.value);
                    context.write(index, result);
                }
            }
        }
    }

    public static void main(String[] args) throws Exception {
        
        Configuration conf = new Configuration();
        
        Job job1 = Job.getInstance(conf, "matix multiplication 1st part");
        
        job1.setJarByClass(MatrixMultiplier.class);
        job1.setMapperClass(MatrixMapper.class);
        job1.setMapOutputValueClass(CellWritable.class);
        //job1.setCombinerClass(IntSumReducer.class);
        job1.setReducerClass(MultiplierReducer.class);
        
        job1.setOutputKeyClass(Text.class);
        job1.setOutputValueClass(IntWritable.class);
        
        TextInputFormat.addInputPath(job1, new Path(args[0]));
        FileOutputFormat.setOutputPath(job1, new Path(args[1] + "_partial"));
 
        if (!job1.waitForCompletion(true)) {
            System.exit(1);
        }
       
        Job job2 = Job.getInstance(conf, "matix multiplication 2nd part");
        
        job2.setJarByClass(MatrixMultiplier.class);
        job2.setMapperClass(MyIdentityMapper.class);
        //job2.setCombinerClass(IntSumReducer.class);
        job2.setReducerClass(IntSumReducer.class);
        
        job2.setOutputKeyClass(Text.class);
        job2.setOutputValueClass(IntWritable.class);
        
        TextInputFormat.addInputPath(job2, new Path(args[1] + "_partial"));
        FileOutputFormat.setOutputPath(job2, new Path(args[1]));
 
        System.exit(job2.waitForCompletion(true) ? 0 : 1);
    }
}

