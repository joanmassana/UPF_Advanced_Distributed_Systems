hadoop com.sun.tools.javac.Main MatrixMultiplier.java
jar cf mm.jar MatrixMultiplier*.class
hadoop fs -rm -f -r /user/u87515/output/mm_partial/
hadoop fs -rm -f -r /user/u87515/output/mm/
hadoop jar mm.jar MatrixMultiplier /user/u87515/input/mm /user/u87515/output/mm
hadoop fs -cat /user/u87515/output/mm/part-r-00000

