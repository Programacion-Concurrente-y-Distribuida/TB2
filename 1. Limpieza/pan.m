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

		 /* CLAIM invariante_dataset */
	case 3: // STATE 1 - _spin_nvr.tmp:32 - [(!((total_records==finalDataset_size)))] (6:0:0 - 1)
		
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
		reached[7][1] = 1;
		if (!( !((((int)now.total_records)==((int)now.finalDataset_size)))))
			continue;
		/* merge: assert(!(!((total_records==finalDataset_size))))(0, 2, 6) */
		reached[7][2] = 1;
		spin_assert( !( !((((int)now.total_records)==((int)now.finalDataset_size)))), " !( !((total_records==finalDataset_size)))", II, tt, t);
		/* merge: .(goto)(0, 7, 6) */
		reached[7][7] = 1;
		;
		_m = 3; goto P999; /* 2 */
	case 4: // STATE 10 - _spin_nvr.tmp:37 - [-end-] (0:0:0 - 1)
		
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
		reached[7][10] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* CLAIM progreso_workers */
	case 5: // STATE 1 - _spin_nvr.tmp:21 - [((!(!((workers_vivos>0)))&&!((workers_vivos==0))))] (0:0:0 - 1)
		
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
		reached[6][1] = 1;
		if (!(( !( !((((int)now.workers_vivos)>0)))&& !((((int)now.workers_vivos)==0)))))
			continue;
		_m = 3; goto P999; /* 0 */
	case 6: // STATE 8 - _spin_nvr.tmp:26 - [(!((workers_vivos==0)))] (0:0:0 - 1)
		
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
		reached[6][8] = 1;
		if (!( !((((int)now.workers_vivos)==0))))
			continue;
		_m = 3; goto P999; /* 0 */
	case 7: // STATE 13 - _spin_nvr.tmp:28 - [-end-] (0:0:0 - 1)
		
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
		reached[6][13] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* CLAIM cada_job_se_procesa */
	case 8: // STATE 1 - _spin_nvr.tmp:10 - [((!(!((jobs_enviados==6)))&&!((jobs_procesados==6))))] (0:0:0 - 1)
		
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
		if (!(( !( !((((int)now.jobs_enviados)==6)))&& !((((int)now.jobs_procesados)==6)))))
			continue;
		_m = 3; goto P999; /* 0 */
	case 9: // STATE 8 - _spin_nvr.tmp:15 - [(!((jobs_procesados==6)))] (0:0:0 - 1)
		
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
		reached[5][8] = 1;
		if (!( !((((int)now.jobs_procesados)==6))))
			continue;
		_m = 3; goto P999; /* 0 */
	case 10: // STATE 13 - _spin_nvr.tmp:17 - [-end-] (0:0:0 - 1)
		
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
		reached[5][13] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* CLAIM todos_procesados */
	case 11: // STATE 1 - _spin_nvr.tmp:4 - [(!((jobs_procesados==6)))] (0:0:0 - 1)
		
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
		if (!( !((((int)now.jobs_procesados)==6))))
			continue;
		_m = 3; goto P999; /* 0 */
	case 12: // STATE 6 - _spin_nvr.tmp:6 - [-end-] (0:0:0 - 1)
		
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
		reached[4][6] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* PROC :init: */
	case 13: // STATE 1 - union_y_limpieza.pml:95 - [((w<3))] (0:0:0 - 1)
		IfNotBlocked
		reached[3][1] = 1;
		if (!((((int)((P3 *)_this)->w)<3)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 14: // STATE 2 - union_y_limpieza.pml:96 - [workers_vivos = (workers_vivos+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[3][2] = 1;
		(trpt+1)->bup.oval = ((int)now.workers_vivos);
		now.workers_vivos = (((int)now.workers_vivos)+1);
#ifdef VAR_RANGES
		logval("workers_vivos", ((int)now.workers_vivos));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 15: // STATE 4 - union_y_limpieza.pml:97 - [(run worker())] (0:0:0 - 1)
		IfNotBlocked
		reached[3][4] = 1;
		if (!(addproc(II, 1, 0)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 16: // STATE 5 - union_y_limpieza.pml:98 - [w = (w+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[3][5] = 1;
		(trpt+1)->bup.oval = ((int)((P3 *)_this)->w);
		((P3 *)_this)->w = (((int)((P3 *)_this)->w)+1);
#ifdef VAR_RANGES
		logval(":init::w", ((int)((P3 *)_this)->w));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 17: // STATE 11 - union_y_limpieza.pml:104 - [((i<6))] (0:0:0 - 1)
		IfNotBlocked
		reached[3][11] = 1;
		if (!((((int)((P3 *)_this)->i)<6)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 18: // STATE 12 - union_y_limpieza.pml:105 - [jobs!i] (0:0:0 - 1)
		IfNotBlocked
		reached[3][12] = 1;
		if (q_full(now.jobs))
			continue;
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[64];
			sprintf(simvals, "%d!", now.jobs);
		sprintf(simtmp, "%d", ((int)((P3 *)_this)->i)); strcat(simvals, simtmp);		}
#endif
		
		qsend(now.jobs, 0, ((int)((P3 *)_this)->i), 1);
		_m = 2; goto P999; /* 0 */
	case 19: // STATE 13 - union_y_limpieza.pml:106 - [jobs_enviados = (jobs_enviados+1)] (0:18:2 - 1)
		IfNotBlocked
		reached[3][13] = 1;
		(trpt+1)->bup.ovals = grab_ints(2);
		(trpt+1)->bup.ovals[0] = ((int)now.jobs_enviados);
		now.jobs_enviados = (((int)now.jobs_enviados)+1);
#ifdef VAR_RANGES
		logval("jobs_enviados", ((int)now.jobs_enviados));
#endif
		;
		/* merge: i = (i+1)(18, 15, 18) */
		reached[3][15] = 1;
		(trpt+1)->bup.ovals[1] = ((int)((P3 *)_this)->i);
		((P3 *)_this)->i = (((int)((P3 *)_this)->i)+1);
#ifdef VAR_RANGES
		logval(":init::i", ((int)((P3 *)_this)->i));
#endif
		;
		/* merge: .(goto)(0, 19, 18) */
		reached[3][19] = 1;
		;
		_m = 3; goto P999; /* 2 */
	case 20: // STATE 21 - union_y_limpieza.pml:110 - [jobs_closed = 1] (0:0:1 - 3)
		IfNotBlocked
		reached[3][21] = 1;
		(trpt+1)->bup.oval = ((int)now.jobs_closed);
		now.jobs_closed = 1;
#ifdef VAR_RANGES
		logval("jobs_closed", ((int)now.jobs_closed));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 21: // STATE 22 - union_y_limpieza.pml:113 - [(run cerrador())] (0:0:0 - 1)
		IfNotBlocked
		reached[3][22] = 1;
		if (!(addproc(II, 1, 1)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 22: // STATE 23 - union_y_limpieza.pml:114 - [(run recolector())] (0:0:0 - 1)
		IfNotBlocked
		reached[3][23] = 1;
		if (!(addproc(II, 1, 2)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 23: // STATE 24 - union_y_limpieza.pml:115 - [-end-] (0:0:0 - 1)
		IfNotBlocked
		reached[3][24] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* PROC recolector */
	case 24: // STATE 1 - union_y_limpieza.pml:70 - [results?filas] (0:0:1 - 1)
		reached[2][1] = 1;
		if (q_len(now.results) == 0) continue;

		XX=1;
		(trpt+1)->bup.oval = ((int)((P2 *)_this)->filas);
		;
		((P2 *)_this)->filas = qrecv(now.results, XX-1, 0, 1);
#ifdef VAR_RANGES
		logval("recolector:filas", ((int)((P2 *)_this)->filas));
#endif
		;
		
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[32];
			sprintf(simvals, "%d?", now.results);
		sprintf(simtmp, "%d", ((int)((P2 *)_this)->filas)); strcat(simvals, simtmp);		}
#endif
		;
		_m = 4; goto P999; /* 0 */
	case 25: // STATE 2 - union_y_limpieza.pml:71 - [mu!1] (0:0:0 - 1)
		IfNotBlocked
		reached[2][2] = 1;
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
	case 26: // STATE 3 - union_y_limpieza.pml:72 - [assert((mutex_ocupado==0))] (0:0:0 - 1)
		IfNotBlocked
		reached[2][3] = 1;
		spin_assert((((int)now.mutex_ocupado)==0), "(mutex_ocupado==0)", II, tt, t);
		_m = 3; goto P999; /* 0 */
	case 27: // STATE 4 - union_y_limpieza.pml:73 - [mutex_ocupado = 1] (0:0:1 - 1)
		IfNotBlocked
		reached[2][4] = 1;
		(trpt+1)->bup.oval = ((int)now.mutex_ocupado);
		now.mutex_ocupado = 1;
#ifdef VAR_RANGES
		logval("mutex_ocupado", ((int)now.mutex_ocupado));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 28: // STATE 5 - union_y_limpieza.pml:74 - [finalDataset_size = (finalDataset_size+filas)] (0:0:1 - 1)
		IfNotBlocked
		reached[2][5] = 1;
		(trpt+1)->bup.oval = ((int)now.finalDataset_size);
		now.finalDataset_size = (((int)now.finalDataset_size)+((int)((P2 *)_this)->filas));
#ifdef VAR_RANGES
		logval("finalDataset_size", ((int)now.finalDataset_size));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 29: // STATE 6 - union_y_limpieza.pml:75 - [total_records = (total_records+filas)] (0:0:1 - 1)
		IfNotBlocked
		reached[2][6] = 1;
		(trpt+1)->bup.oval = ((int)now.total_records);
		now.total_records = (((int)now.total_records)+((int)((P2 *)_this)->filas));
#ifdef VAR_RANGES
		logval("total_records", ((int)now.total_records));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 30: // STATE 7 - union_y_limpieza.pml:76 - [mutex_ocupado = 0] (0:0:1 - 1)
		IfNotBlocked
		reached[2][7] = 1;
		(trpt+1)->bup.oval = ((int)now.mutex_ocupado);
		now.mutex_ocupado = 0;
#ifdef VAR_RANGES
		logval("mutex_ocupado", ((int)now.mutex_ocupado));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 31: // STATE 8 - union_y_limpieza.pml:77 - [mu?1] (0:0:0 - 1)
		reached[2][8] = 1;
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
	case 32: // STATE 9 - union_y_limpieza.pml:78 - [((empty(results)&&results_closed))] (0:0:0 - 1)
		IfNotBlocked
		reached[2][9] = 1;
		if (!(((q_len(now.results)==0)&&((int)now.results_closed))))
			continue;
		_m = 3; goto P999; /* 0 */
	case 33: // STATE 14 - union_y_limpieza.pml:82 - [assert((jobs_enviados==6))] (0:0:0 - 3)
		IfNotBlocked
		reached[2][14] = 1;
		spin_assert((((int)now.jobs_enviados)==6), "(jobs_enviados==6)", II, tt, t);
		_m = 3; goto P999; /* 0 */
	case 34: // STATE 15 - union_y_limpieza.pml:83 - [assert((jobs_procesados==6))] (0:0:0 - 1)
		IfNotBlocked
		reached[2][15] = 1;
		spin_assert((((int)now.jobs_procesados)==6), "(jobs_procesados==6)", II, tt, t);
		_m = 3; goto P999; /* 0 */
	case 35: // STATE 16 - union_y_limpieza.pml:84 - [assert((total_records==finalDataset_size))] (0:0:0 - 1)
		IfNotBlocked
		reached[2][16] = 1;
		spin_assert((((int)now.total_records)==((int)now.finalDataset_size)), "(total_records==finalDataset_size)", II, tt, t);
		_m = 3; goto P999; /* 0 */
	case 36: // STATE 17 - union_y_limpieza.pml:85 - [-end-] (0:0:0 - 1)
		IfNotBlocked
		reached[2][17] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* PROC cerrador */
	case 37: // STATE 1 - union_y_limpieza.pml:60 - [((workers_vivos==0))] (0:0:0 - 1)
		IfNotBlocked
		reached[1][1] = 1;
		if (!((((int)now.workers_vivos)==0)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 38: // STATE 2 - union_y_limpieza.pml:61 - [results_closed = 1] (0:0:1 - 1)
		IfNotBlocked
		reached[1][2] = 1;
		(trpt+1)->bup.oval = ((int)now.results_closed);
		now.results_closed = 1;
#ifdef VAR_RANGES
		logval("results_closed", ((int)now.results_closed));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 39: // STATE 3 - union_y_limpieza.pml:62 - [-end-] (0:0:0 - 1)
		IfNotBlocked
		reached[1][3] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* PROC worker */
	case 40: // STATE 1 - union_y_limpieza.pml:39 - [jobs?archivo] (0:0:2 - 1)
		reached[0][1] = 1;
		if (q_len(now.jobs) == 0) continue;

		XX=1;
		(trpt+1)->bup.ovals = grab_ints(2);
		(trpt+1)->bup.ovals[0] = ((int)((P0 *)_this)->archivo);
		;
		((P0 *)_this)->archivo = qrecv(now.jobs, XX-1, 0, 1);
#ifdef VAR_RANGES
		logval("worker:archivo", ((int)((P0 *)_this)->archivo));
#endif
		;
		
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[32];
			sprintf(simvals, "%d?", now.jobs);
		sprintf(simtmp, "%d", ((int)((P0 *)_this)->archivo)); strcat(simvals, simtmp);		}
#endif
		;
		if (TstOnly) return 1; /* TT */
		/* dead 2: archivo */  (trpt+1)->bup.ovals[1] = ((P0 *)_this)->archivo;
#ifdef HAS_CODE
		if (!readtrail)
#endif
			((P0 *)_this)->archivo = 0;
		_m = 4; goto P999; /* 0 */
	case 41: // STATE 2 - union_y_limpieza.pml:44 - [filas = 0] (0:0:1 - 1)
		IfNotBlocked
		reached[0][2] = 1;
		(trpt+1)->bup.oval = ((int)((P0 *)_this)->filas);
		((P0 *)_this)->filas = 0;
#ifdef VAR_RANGES
		logval("worker:filas", ((int)((P0 *)_this)->filas));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 42: // STATE 3 - union_y_limpieza.pml:45 - [filas = 1] (0:0:1 - 1)
		IfNotBlocked
		reached[0][3] = 1;
		(trpt+1)->bup.oval = ((int)((P0 *)_this)->filas);
		((P0 *)_this)->filas = 1;
#ifdef VAR_RANGES
		logval("worker:filas", ((int)((P0 *)_this)->filas));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 43: // STATE 4 - union_y_limpieza.pml:46 - [filas = 2] (0:0:1 - 1)
		IfNotBlocked
		reached[0][4] = 1;
		(trpt+1)->bup.oval = ((int)((P0 *)_this)->filas);
		((P0 *)_this)->filas = 2;
#ifdef VAR_RANGES
		logval("worker:filas", ((int)((P0 *)_this)->filas));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 44: // STATE 5 - union_y_limpieza.pml:47 - [filas = 2] (0:0:1 - 1)
		IfNotBlocked
		reached[0][5] = 1;
		(trpt+1)->bup.oval = ((int)((P0 *)_this)->filas);
		((P0 *)_this)->filas = 2;
#ifdef VAR_RANGES
		logval("worker:filas", ((int)((P0 *)_this)->filas));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 45: // STATE 8 - union_y_limpieza.pml:49 - [results!filas] (0:0:0 - 5)
		IfNotBlocked
		reached[0][8] = 1;
		if (q_full(now.results))
			continue;
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[64];
			sprintf(simvals, "%d!", now.results);
		sprintf(simtmp, "%d", ((int)((P0 *)_this)->filas)); strcat(simvals, simtmp);		}
#endif
		
		qsend(now.results, 0, ((int)((P0 *)_this)->filas), 1);
		_m = 2; goto P999; /* 0 */
	case 46: // STATE 9 - union_y_limpieza.pml:50 - [jobs_procesados = (jobs_procesados+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[0][9] = 1;
		(trpt+1)->bup.oval = ((int)now.jobs_procesados);
		now.jobs_procesados = (((int)now.jobs_procesados)+1);
#ifdef VAR_RANGES
		logval("jobs_procesados", ((int)now.jobs_procesados));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 47: // STATE 11 - union_y_limpieza.pml:51 - [((empty(jobs)&&jobs_closed))] (0:0:0 - 1)
		IfNotBlocked
		reached[0][11] = 1;
		if (!(((q_len(now.jobs)==0)&&((int)now.jobs_closed))))
			continue;
		_m = 3; goto P999; /* 0 */
	case 48: // STATE 16 - union_y_limpieza.pml:54 - [workers_vivos = (workers_vivos-1)] (0:0:1 - 1)
		IfNotBlocked
		reached[0][16] = 1;
		(trpt+1)->bup.oval = ((int)now.workers_vivos);
		now.workers_vivos = (((int)now.workers_vivos)-1);
#ifdef VAR_RANGES
		logval("workers_vivos", ((int)now.workers_vivos));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 49: // STATE 18 - union_y_limpieza.pml:55 - [-end-] (0:0:0 - 1)
		IfNotBlocked
		reached[0][18] = 1;
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

