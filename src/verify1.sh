cp testcases/testcase1/input-0.dat INPUT 
cat testcases/testcase1/input-1.dat >> INPUT
cat testcases/testcase1/input-2.dat >> INPUT
cat testcases/testcase1/input-3.dat >> INPUT
cat testcases/testcase1/input-4.dat >> INPUT
cat testcases/testcase1/input-5.dat >> INPUT
cat testcases/testcase1/input-6.dat >> INPUT
cat testcases/testcase1/input-7.dat >> INPUT
cat testcases/testcase1/input-8.dat >> INPUT
cat testcases/testcase1/input-9.dat >> INPUT
cat testcases/testcase1/input-10.dat >> INPUT
cat testcases/testcase1/input-11.dat >> INPUT
cat testcases/testcase1/input-12.dat >> INPUT
cat testcases/testcase1/input-13.dat >> INPUT
cat testcases/testcase1/input-14.dat >> INPUT
cat testcases/testcase1/input-15.dat >> INPUT

cp testcases/testcase1/output-0.dat OUTPUT 
cat testcases/testcase1/output-1.dat >> OUTPUT
cat testcases/testcase1/output-2.dat >> OUTPUT
cat testcases/testcase1/output-3.dat >> OUTPUT
cat testcases/testcase1/output-4.dat >> OUTPUT
cat testcases/testcase1/output-5.dat >> OUTPUT
cat testcases/testcase1/output-6.dat >> OUTPUT
cat testcases/testcase1/output-7.dat >> OUTPUT
cat testcases/testcase1/output-8.dat >> OUTPUT
cat testcases/testcase1/output-9.dat >> OUTPUT
cat testcases/testcase1/output-10.dat >> OUTPUT
cat testcases/testcase1/output-11.dat >> OUTPUT
cat testcases/testcase1/output-12.dat >> OUTPUT
cat testcases/testcase1/output-13.dat >> OUTPUT
cat testcases/testcase1/output-14.dat >> OUTPUT
cat testcases/testcase1/output-15.dat >> OUTPUT

./utils/mac-intel/bin/showsort INPUT | sort > REF_OUTPUT
./utils/mac-intel/bin/showsort OUTPUT > my_output

diff REF_OUTPUT my_output

