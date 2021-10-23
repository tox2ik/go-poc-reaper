#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>
#include <limits.h>
#include <stdlib.h>
#include <sys/wait.h>
#include <pthread.h>
#include <sys/prctl.h>

#define RESET   "\033[0m"
#define RESETn   "\033[0m\n"
#define BLACK   "\033[30m"      /* Black */
#define RED     "\033[31m"      /* Red */
#define GREEN   "\033[32m"      /* Green */
#define YELLOW  "\033[33m"      /* Yellow */
#define BLUE    "\033[34m"      /* Blue */
#define MAGENTA "\033[35m"      /* Magenta */
#define CYAN    "\033[36m"      /* Cyan */
#define WHITE   "\033[37m"      /* White */
#define BOLDBLACK   "\033[1m\033[30m"      /* Bold Black */
#define BOLDRED     "\033[1m\033[31m"      /* Bold Red */
#define BOLDGREEN   "\033[1m\033[32m"      /* Bold Green */
#define BOLDYELLOW  "\033[1m\033[33m"      /* Bold Yellow */
#define BOLDBLUE    "\033[1m\033[34m"      /* Bold Blue */
#define BOLDMAGENTA "\033[1m\033[35m"      /* Bold Magenta */
#define BOLDCYAN    "\033[1m\033[36m"      /* Bold Cyan */
#define BOLDWHITE   "\033[1m\033[37m"      /* Bold White */


void forever() {
	printf(BOLDBLUE "daemon 1" RESETn, "");
	pthread_setname_np(pthread_self(), "fork-daemon");
	while (1) {
			sleep(1);
			printf(RED "daemon" RESETn);
	}
}
void child(int x) {
	for (int j=x; j > 0 ; j--) {
		for (int i=0; i < INT_MAX/5; i++ ) {
			printf("");
		}
	}
	printf(RESET);
    printf(RED "Child '%d' done" RESETn, x);
}

void severalForks() {
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
					printf(RED "Hello from Parent! %7d %7d %7d " BOLDBLUE "%7d" RESETn, p1, p2, p3, d);
				}
			} else {
				printf(RED "Hello from Parent! %7d %7d %7d %7d" RESETn, p1, p2, p3, d);
				child(6);
			}
		} else {
			printf(RED "Hello from Parent! %7d %7d %7d %7d" RESETn, p1, p2, p3, d);
			child(2);
		}
	} else {
        printf(RED "Hello from Parent! %7d %7d %7d %7d" RESETn, p1, p2, p3, d);
		child(1);
	}

	if (d > 0) {
		printf(BOLDWHITE "fork.main() " BOLDBLUE "wait! " BOLDWHITE "%d %d %d " BOLDBLUE "%d" RESETn, p1, p2, p3, d);
		wait4(d, 0, 0, 0);
		printf(BOLDWHITE "fork.main() " BOLDBLUE "daemon done." RESETn);
	} else {
		printf((BOLDGREEN "thread done" RESETn));
	}
}


int main() {
    severalForks();
    return 0;
}
