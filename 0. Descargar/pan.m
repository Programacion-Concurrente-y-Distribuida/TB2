#define rand	pan_rand
#define pthread_equal(a,b)	((a)==(b))
#if defined(HAS_CODE) && defined(VERBOSE)
	#ifdef BFS_PAR
		bfs_printf("Pr: %d Tr: %d\n", II, t->forw);
	#else
		cpu_printf("Pr: %d Tr: %d\n", II, t->forw);
	#endif
#endif
	switch (t->forw) {
	default: Uerror("bad forward move");
	case 0:	/* if without executable clauses */
		continue;
	case 1: /* generic 'goto' or 'skip' */
		IfNotBlocked
		_m = 3; goto P999;
	case 2: /* generic 'else' */
		IfNotBlocked
		if (trpt->o_pm&1) continue;
		_m = 3; goto P999;

		 /* CLAIM sem_respetado */
	case 3: // STATE 1 - _spin_nvr.tmp:30 - [(!((downloaded_count<=10)))] (6:0:0 - 1)
		
#if defined(VERI) && !defined(NP)
#if NCLAIMS>1
		{	static int reported1 = 0;
			if (verbose && !reported1)
			{	int nn = (int) ((Pclaim *)pptr(0))->_n;
				printf("depth %ld: Claim %s (%d), state %d (line %d)\n",
					depth, procname[spin_c_typ[nn]], nn, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported1 = 1;
				fflush(stdout);
		}	}
#else
		{	static int reported1 = 0;
			if (verbose && !reported1)
			{	printf("depth %d: Claim, state %d (line %d)\n",
					(int) depth, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported1 = 1;
				fflush(stdout);
		}	}
#endif
#endif
		reached[5][1] = 1;
		if (!( !((((int)now.downloaded_count)<=10))))
			continue;
		/* merge: assert(!(!((downloaded_count<=10))))(0, 2, 6) */
		reached[5][2] = 1;
		spin_assert( !( !((((int)now.downloaded_count)<=10))), " !( !((downloaded_count<=10)))", II, tt, t);
		/* merge: .(goto)(0, 7, 6) */
		reached[5][7] = 1;
		;
		_m = 3; goto P999; /* 2 */
	case 4: // STATE 10 - _spin_nvr.tmp:35 - [-end-] (0:0:0 - 1)
		
#if defined(VERI) && !defined(NP)
#if NCLAIMS>1
		{	static int reported10 = 0;
			if (verbose && !reported10)
			{	int nn = (int) ((Pclaim *)pptr(0))->_n;
				printf("depth %ld: Claim %s (%d), state %d (line %d)\n",
					depth, procname[spin_c_typ[nn]], nn, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported10 = 1;
				fflush(stdout);
		}	}
#else
		{	static int reported10 = 0;
			if (verbose && !reported10)
			{	printf("depth %d: Claim, state %d (line %d)\n",
					(int) depth, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported10 = 1;
				fflush(stdout);
		}	}
#endif
#endif
		reached[5][10] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* CLAIM progreso_wg */
	case 5: // STATE 1 - _spin_nvr.tmp:19 - [((!(!((active_downloads>0)))&&!((active_downloads==0))))] (0:0:0 - 1)
		
#if defined(VERI) && !defined(NP)
#if NCLAIMS>1
		{	static int reported1 = 0;
			if (verbose && !reported1)
			{	int nn = (int) ((Pclaim *)pptr(0))->_n;
				printf("depth %ld: Claim %s (%d), state %d (line %d)\n",
					depth, procname[spin_c_typ[nn]], nn, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported1 = 1;
				fflush(stdout);
		}	}
#else
		{	static int reported1 = 0;
			if (verbose && !reported1)
			{	printf("depth %d: Claim, state %d (line %d)\n",
					(int) depth, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported1 = 1;
				fflush(stdout);
		}	}
#endif
#endif
		reached[4][1] = 1;
		if (!(( !( !((((int)now.active_downloads)>0)))&& !((((int)now.active_downloads)==0)))))
			continue;
		_m = 3; goto P999; /* 0 */
	case 6: // STATE 8 - _spin_nvr.tmp:24 - [(!((active_downloads==0)))] (0:0:0 - 1)
		
#if defined(VERI) && !defined(NP)
#if NCLAIMS>1
		{	static int reported8 = 0;
			if (verbose && !reported8)
			{	int nn = (int) ((Pclaim *)pptr(0))->_n;
				printf("depth %ld: Claim %s (%d), state %d (line %d)\n",
					depth, procname[spin_c_typ[nn]], nn, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported8 = 1;
				fflush(stdout);
		}	}
#else
		{	static int reported8 = 0;
			if (verbose && !reported8)
			{	printf("depth %d: Claim, state %d (line %d)\n",
					(int) depth, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported8 = 1;
				fflush(stdout);
		}	}
#endif
#endif
		reached[4][8] = 1;
		if (!( !((((int)now.active_downloads)==0))))
			continue;
		_m = 3; goto P999; /* 0 */
	case 7: // STATE 13 - _spin_nvr.tmp:26 - [-end-] (0:0:0 - 1)
		
#if defined(VERI) && !defined(NP)
#if NCLAIMS>1
		{	static int reported13 = 0;
			if (verbose && !reported13)
			{	int nn = (int) ((Pclaim *)pptr(0))->_n;
				printf("depth %ld: Claim %s (%d), state %d (line %d)\n",
					depth, procname[spin_c_typ[nn]], nn, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported13 = 1;
				fflush(stdout);
		}	}
#else
		{	static int reported13 = 0;
			if (verbose && !reported13)
			{	printf("depth %d: Claim, state %d (line %d)\n",
					(int) depth, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported13 = 1;
				fflush(stdout);
		}	}
#endif
#endif
		reached[4][13] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* CLAIM terminacion */
	case 8: // STATE 1 - _spin_nvr.tmp:13 - [(!((finished_count==10)))] (0:0:0 - 1)
		
#if defined(VERI) && !defined(NP)
#if NCLAIMS>1
		{	static int reported1 = 0;
			if (verbose && !reported1)
			{	int nn = (int) ((Pclaim *)pptr(0))->_n;
				printf("depth %ld: Claim %s (%d), state %d (line %d)\n",
					depth, procname[spin_c_typ[nn]], nn, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported1 = 1;
				fflush(stdout);
		}	}
#else
		{	static int reported1 = 0;
			if (verbose && !reported1)
			{	printf("depth %d: Claim, state %d (line %d)\n",
					(int) depth, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported1 = 1;
				fflush(stdout);
		}	}
#endif
#endif
		reached[3][1] = 1;
		if (!( !((((int)now.finished_count)==10))))
			continue;
		_m = 3; goto P999; /* 0 */
	case 9: // STATE 6 - _spin_nvr.tmp:15 - [-end-] (0:0:0 - 1)
		
#if defined(VERI) && !defined(NP)
#if NCLAIMS>1
		{	static int reported6 = 0;
			if (verbose && !reported6)
			{	int nn = (int) ((Pclaim *)pptr(0))->_n;
				printf("depth %ld: Claim %s (%d), state %d (line %d)\n",
					depth, procname[spin_c_typ[nn]], nn, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported6 = 1;
				fflush(stdout);
		}	}
#else
		{	static int reported6 = 0;
			if (verbose && !reported6)
			{	printf("depth %d: Claim, state %d (line %d)\n",
					(int) depth, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported6 = 1;
				fflush(stdout);
		}	}
#endif
#endif
		reached[3][6] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* CLAIM no_exceso */
	case 10: // STATE 1 - _spin_nvr.tmp:3 - [(!((downloaded_count<=10)))] (6:0:0 - 1)
		
#if defined(VERI) && !defined(NP)
#if NCLAIMS>1
		{	static int reported1 = 0;
			if (verbose && !reported1)
			{	int nn = (int) ((Pclaim *)pptr(0))->_n;
				printf("depth %ld: Claim %s (%d), state %d (line %d)\n",
					depth, procname[spin_c_typ[nn]], nn, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported1 = 1;
				fflush(stdout);
		}	}
#else
		{	static int reported1 = 0;
			if (verbose && !reported1)
			{	printf("depth %d: Claim, state %d (line %d)\n",
					(int) depth, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported1 = 1;
				fflush(stdout);
		}	}
#endif
#endif
		reached[2][1] = 1;
		if (!( !((((int)now.downloaded_count)<=10))))
			continue;
		/* merge: assert(!(!((downloaded_count<=10))))(0, 2, 6) */
		reached[2][2] = 1;
		spin_assert( !( !((((int)now.downloaded_count)<=10))), " !( !((downloaded_count<=10)))", II, tt, t);
		/* merge: .(goto)(0, 7, 6) */
		reached[2][7] = 1;
		;
		_m = 3; goto P999; /* 2 */
	case 11: // STATE 10 - _spin_nvr.tmp:8 - [-end-] (0:0:0 - 1)
		
#if defined(VERI) && !defined(NP)
#if NCLAIMS>1
		{	static int reported10 = 0;
			if (verbose && !reported10)
			{	int nn = (int) ((Pclaim *)pptr(0))->_n;
				printf("depth %ld: Claim %s (%d), state %d (line %d)\n",
					depth, procname[spin_c_typ[nn]], nn, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported10 = 1;
				fflush(stdout);
		}	}
#else
		{	static int reported10 = 0;
			if (verbose && !reported10)
			{	printf("depth %d: Claim, state %d (line %d)\n",
					(int) depth, (int) ((Pclaim *)pptr(0))->_p, src_claim[ (int) ((Pclaim *)pptr(0))->_p ]);
				reported10 = 1;
				fflush(stdout);
		}	}
#endif
#endif
		reached[2][10] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* PROC :init: */
	case 12: // STATE 1 - descargar.pml:62 - [((i<10))] (0:0:0 - 1)
		IfNotBlocked
		reached[1][1] = 1;
		if (!((((int)((P1 *)_this)->i)<10)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 13: // STATE 2 - descargar.pml:63 - [sem!1] (0:0:0 - 1)
		IfNotBlocked
		reached[1][2] = 1;
		if (q_full(now.sem))
			continue;
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[64];
			sprintf(simvals, "%d!", now.sem);
		sprintf(simtmp, "%d", 1); strcat(simvals, simtmp);		}
#endif
		
		qsend(now.sem, 0, 1, 1);
		_m = 2; goto P999; /* 0 */
	case 14: // STATE 3 - descargar.pml:64 - [active_downloads = (active_downloads+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[1][3] = 1;
		(trpt+1)->bup.oval = ((int)now.active_downloads);
		now.active_downloads = (((int)now.active_downloads)+1);
#ifdef VAR_RANGES
		logval("active_downloads", ((int)now.active_downloads));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 15: // STATE 5 - descargar.pml:65 - [(run descargador())] (0:0:0 - 1)
		IfNotBlocked
		reached[1][5] = 1;
		if (!(addproc(II, 1, 0)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 16: // STATE 6 - descargar.pml:66 - [i = (i+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[1][6] = 1;
		(trpt+1)->bup.oval = ((int)((P1 *)_this)->i);
		((P1 *)_this)->i = (((int)((P1 *)_this)->i)+1);
#ifdef VAR_RANGES
		logval(":init::i", ((int)((P1 *)_this)->i));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 17: // STATE 12 - descargar.pml:71 - [((active_downloads==0))] (0:0:0 - 3)
		IfNotBlocked
		reached[1][12] = 1;
		if (!((((int)now.active_downloads)==0)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 18: // STATE 13 - descargar.pml:74 - [assert((finished_count==10))] (0:0:0 - 1)
		IfNotBlocked
		reached[1][13] = 1;
		spin_assert((((int)now.finished_count)==10), "(finished_count==10)", II, tt, t);
		_m = 3; goto P999; /* 0 */
	case 19: // STATE 14 - descargar.pml:77 - [assert((downloaded_count<=10))] (0:0:0 - 1)
		IfNotBlocked
		reached[1][14] = 1;
		spin_assert((((int)now.downloaded_count)<=10), "(downloaded_count<=10)", II, tt, t);
		_m = 3; goto P999; /* 0 */
	case 20: // STATE 15 - descargar.pml:78 - [-end-] (0:0:0 - 1)
		IfNotBlocked
		reached[1][15] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* PROC descargador */
	case 21: // STATE 1 - descargar.pml:31 - [exito = 1] (0:0:1 - 1)
		IfNotBlocked
		reached[0][1] = 1;
		(trpt+1)->bup.oval = ((int)((P0 *)_this)->exito);
		((P0 *)_this)->exito = 1;
#ifdef VAR_RANGES
		logval("descargador:exito", ((int)((P0 *)_this)->exito));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 22: // STATE 2 - descargar.pml:32 - [exito = 0] (0:0:1 - 1)
		IfNotBlocked
		reached[0][2] = 1;
		(trpt+1)->bup.oval = ((int)((P0 *)_this)->exito);
		((P0 *)_this)->exito = 0;
#ifdef VAR_RANGES
		logval("descargador:exito", ((int)((P0 *)_this)->exito));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 23: // STATE 5 - descargar.pml:36 - [((exito==1))] (0:0:1 - 1)
		IfNotBlocked
		reached[0][5] = 1;
		if (!((((int)((P0 *)_this)->exito)==1)))
			continue;
		if (TstOnly) return 1; /* TT */
		/* dead 1: exito */  (trpt+1)->bup.oval = ((P0 *)_this)->exito;
#ifdef HAS_CODE
		if (!readtrail)
#endif
			((P0 *)_this)->exito = 0;
		_m = 3; goto P999; /* 0 */
	case 24: // STATE 6 - descargar.pml:37 - [mu!1] (0:0:0 - 1)
		IfNotBlocked
		reached[0][6] = 1;
		if (q_full(now.mu))
			continue;
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[64];
			sprintf(simvals, "%d!", now.mu);
		sprintf(simtmp, "%d", 1); strcat(simvals, simtmp);		}
#endif
		
		qsend(now.mu, 0, 1, 1);
		_m = 2; goto P999; /* 0 */
	case 25: // STATE 7 - descargar.pml:38 - [assert((mutex_ocupado==0))] (0:0:0 - 1)
		IfNotBlocked
		reached[0][7] = 1;
		spin_assert((((int)now.mutex_ocupado)==0), "(mutex_ocupado==0)", II, tt, t);
		_m = 3; goto P999; /* 0 */
	case 26: // STATE 8 - descargar.pml:39 - [mutex_ocupado = 1] (0:0:1 - 1)
		IfNotBlocked
		reached[0][8] = 1;
		(trpt+1)->bup.oval = ((int)now.mutex_ocupado);
		now.mutex_ocupado = 1;
#ifdef VAR_RANGES
		logval("mutex_ocupado", ((int)now.mutex_ocupado));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 27: // STATE 9 - descargar.pml:40 - [assert((downloaded_count<10))] (0:0:0 - 1)
		IfNotBlocked
		reached[0][9] = 1;
		spin_assert((((int)now.downloaded_count)<10), "(downloaded_count<10)", II, tt, t);
		_m = 3; goto P999; /* 0 */
	case 28: // STATE 10 - descargar.pml:41 - [downloaded_count = (downloaded_count+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[0][10] = 1;
		(trpt+1)->bup.oval = ((int)now.downloaded_count);
		now.downloaded_count = (((int)now.downloaded_count)+1);
#ifdef VAR_RANGES
		logval("downloaded_count", ((int)now.downloaded_count));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 29: // STATE 11 - descargar.pml:42 - [mutex_ocupado = 0] (0:0:1 - 1)
		IfNotBlocked
		reached[0][11] = 1;
		(trpt+1)->bup.oval = ((int)now.mutex_ocupado);
		now.mutex_ocupado = 0;
#ifdef VAR_RANGES
		logval("mutex_ocupado", ((int)now.mutex_ocupado));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 30: // STATE 12 - descargar.pml:43 - [mu?1] (0:0:0 - 1)
		reached[0][12] = 1;
		if (q_len(now.mu) == 0) continue;

		XX=1;
		if (1 != qrecv(now.mu, 0, 0, 0)) continue;
		
#ifndef BFS_PAR
		if (q_flds[((Q0 *)qptr(now.mu-1))->_t] != 1)
			Uerror("wrong nr of msg fields in rcv");
#endif
		;
		qrecv(now.mu, XX-1, 0, 1);
		
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[32];
			sprintf(simvals, "%d?", now.mu);
		sprintf(simtmp, "%d", 1); strcat(simvals, simtmp);		}
#endif
		;
		_m = 4; goto P999; /* 0 */
	case 31: // STATE 17 - descargar.pml:49 - [finished_count = (finished_count+1)] (0:20:2 - 1)
		IfNotBlocked
		reached[0][17] = 1;
		(trpt+1)->bup.ovals = grab_ints(2);
		(trpt+1)->bup.ovals[0] = ((int)now.finished_count);
		now.finished_count = (((int)now.finished_count)+1);
#ifdef VAR_RANGES
		logval("finished_count", ((int)now.finished_count));
#endif
		;
		/* merge: active_downloads = (active_downloads-1)(20, 18, 20) */
		reached[0][18] = 1;
		(trpt+1)->bup.ovals[1] = ((int)now.active_downloads);
		now.active_downloads = (((int)now.active_downloads)-1);
#ifdef VAR_RANGES
		logval("active_downloads", ((int)now.active_downloads));
#endif
		;
		_m = 3; goto P999; /* 1 */
	case 32: // STATE 20 - descargar.pml:53 - [sem?1] (0:0:0 - 1)
		reached[0][20] = 1;
		if (q_len(now.sem) == 0) continue;

		XX=1;
		if (1 != qrecv(now.sem, 0, 0, 0)) continue;
		
#ifndef BFS_PAR
		if (q_flds[((Q0 *)qptr(now.sem-1))->_t] != 1)
			Uerror("wrong nr of msg fields in rcv");
#endif
		;
		qrecv(now.sem, XX-1, 0, 1);
		
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[32];
			sprintf(simvals, "%d?", now.sem);
		sprintf(simtmp, "%d", 1); strcat(simvals, simtmp);		}
#endif
		;
		_m = 4; goto P999; /* 0 */
	case 33: // STATE 21 - descargar.pml:54 - [-end-] (0:0:0 - 1)
		IfNotBlocked
		reached[0][21] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */
	case  _T5:	/* np_ */
		if (!((!(trpt->o_pm&4) && !(trpt->tau&128))))
			continue;
		/* else fall through */
	case  _T2:	/* true */
		_m = 3; goto P999;
#undef rand
	}

