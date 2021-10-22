#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>
#include <limits.h>
#include <stdlib.h>
#include <sys/wait.h>
#include <pthread.h>
#include <sys/prctl.h>

void forever() {
	printf("daemon 1\n");
	pthread_setname_np(pthread_self(), "fork-daemon");
	while (1) {
			sleep(1);
			printf("daemon\n");
	}
}
void child(int x) {
	for (int j=x; j > 0 ; j--) {
		for (int i=0; i < INT_MAX/10; i++ ) {
			printf("");
		}
	}

    printf("Child %d, bye!\n", x);
}

void forkexample()
{
	int d  = -3 ;
	int p1 = -3 ;
	int p2 = -3 ;
	int p3 = -3 ;


    if (( p1 = fork()) == 0) {
		pthread_setname_np(pthread_self(), "fork-child-A");
        // child process because return value zero
		if ((p2=fork()) == 0) {
			pthread_setname_np(pthread_self(), "fork-child-B");
			if ((p3=fork()) == 0) {
				pthread_setname_np(pthread_self(), "fork-child-C");
				if ((d=fork()) == 0) {
					forever();
				} else {
					printf("Hello from Parent! %7d %7d %7d %7d \n", p1, p2, p3, d);
				}
			} else {
				printf("Hello from Parent! %7d %7d %7d %7d \n", p1, p2, p3, d);
				child(6);
			}
		} else {
			printf("Hello from Parent! %7d %7d %7d %7d \n", p1, p2, p3, d);
			child(2);
		}
	} else {
        printf("Hello from Parent! %7d %7d %7d %7d \n", p1, p2, p3, d);
		child(1);
	}

	if (d > 0) {
		printf("fork.main() wait! %d %d %d %d\n", p1, p2, p3, d);
		wait4(d, 0, 0, 0);
		printf("fork.main() done.\n");
	} else {
		printf("thread done\n");
	}
}


int main()
{
    forkexample();
    return 0;
}
