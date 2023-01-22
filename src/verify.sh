rm -rf testcases/testcase1/output-*.dat
rm INPUT 
rm OUTPUT 
rm REF_OUTPUT
rm my_output
./run-demo-4.sh
N=3
cp testcases/testcase1/input-0.dat INPUT 
for i in $(seq 1 ${N})
do
    # echo testcases/testcase1/input-${i}.dat
    cat testcases/testcase1/input-${i}.dat >> INPUT
done 
wait 
cp testcases/testcase1/output-0.dat OUTPUT 
for i in $(seq 1 ${N})
do
    # echo testcases/testcase1/output-${i}.dat
    cat testcases/testcase1/output-${i}.dat >> OUTPUT
done 
wait

./utils/mac-intel/bin/showsort INPUT | sort > REF_OUTPUT
./utils/mac-intel/bin/showsort OUTPUT > my_output

diff REF_OUTPUT my_output

