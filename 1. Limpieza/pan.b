	switch (t->back) {
	default: Uerror("bad return move");
	case  0: goto R999; /* nothing to undo */

		 /* CLAIM invariante_dataset */
;
		
	case 3: // STATE 1
		goto R999;

	case 4: // STATE 10
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* CLAIM progreso_workers */
;
		;
		;
		;
		
	case 7: // STATE 13
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* CLAIM cada_job_se_procesa */
;
		;
		;
		;
		
	case 10: // STATE 13
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* CLAIM todos_procesados */
;
		;
		
	case 12: // STATE 6
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* PROC :init: */
;
		;
		
	case 14: // STATE 2
		;
		now.workers_vivos = trpt->bup.oval;
		;
		goto R999;

	case 15: // STATE 4
		;
		;
		delproc(0, now._nr_pr-1);
		;
		goto R999;

	case 16: // STATE 5
		;
		((P3 *)_this)->w = trpt->bup.oval;
		;
		goto R999;
;
		;
		
	case 18: // STATE 12
		;
		_m = unsend(now.jobs);
		;
		goto R999;

	case 19: // STATE 15
		;
		((P3 *)_this)->i = trpt->bup.ovals[1];
		now.jobs_enviados = trpt->bup.ovals[0];
		;
		ungrab_ints(trpt->bup.ovals, 2);
		goto R999;

	case 20: // STATE 21
		;
		now.jobs_closed = trpt->bup.oval;
		;
		goto R999;

	case 21: // STATE 22
		;
		;
		delproc(0, now._nr_pr-1);
		;
		goto R999;

	case 22: // STATE 23
		;
		;
		delproc(0, now._nr_pr-1);
		;
		goto R999;

	case 23: // STATE 24
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* PROC recolector */

	case 24: // STATE 1
		;
		XX = 1;
		unrecv(now.results, XX-1, 0, ((int)((P2 *)_this)->filas), 1);
		((P2 *)_this)->filas = trpt->bup.oval;
		;
		;
		goto R999;

	case 25: // STATE 2
		;
		_m = unsend(now.mu);
		;
		goto R999;
;
		;
		
	case 27: // STATE 4
		;
		now.mutex_ocupado = trpt->bup.oval;
		;
		goto R999;

	case 28: // STATE 5
		;
		now.finalDataset_size = trpt->bup.oval;
		;
		goto R999;

	case 29: // STATE 6
		;
		now.total_records = trpt->bup.oval;
		;
		goto R999;

	case 30: // STATE 7
		;
		now.mutex_ocupado = trpt->bup.oval;
		;
		goto R999;

	case 31: // STATE 8
		;
		XX = 1;
		unrecv(now.mu, XX-1, 0, 1, 1);
		;
		;
		goto R999;
;
		;
		;
		;
		;
		;
		;
		;
		
	case 36: // STATE 17
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* PROC cerrador */
;
		;
		
	case 38: // STATE 2
		;
		now.results_closed = trpt->bup.oval;
		;
		goto R999;

	case 39: // STATE 3
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* PROC worker */

	case 40: // STATE 1
		;
	/* 0 */	((P0 *)_this)->archivo = trpt->bup.ovals[1];
		XX = 1;
		unrecv(now.jobs, XX-1, 0, ((int)((P0 *)_this)->archivo), 1);
		((P0 *)_this)->archivo = trpt->bup.ovals[0];
		;
		;
		ungrab_ints(trpt->bup.ovals, 2);
		goto R999;

	case 41: // STATE 2
		;
		((P0 *)_this)->filas = trpt->bup.oval;
		;
		goto R999;

	case 42: // STATE 3
		;
		((P0 *)_this)->filas = trpt->bup.oval;
		;
		goto R999;

	case 43: // STATE 4
		;
		((P0 *)_this)->filas = trpt->bup.oval;
		;
		goto R999;

	case 44: // STATE 5
		;
		((P0 *)_this)->filas = trpt->bup.oval;
		;
		goto R999;

	case 45: // STATE 8
		;
		_m = unsend(now.results);
		;
		goto R999;

	case 46: // STATE 9
		;
		now.jobs_procesados = trpt->bup.oval;
		;
		goto R999;
;
		;
		
	case 48: // STATE 16
		;
		now.workers_vivos = trpt->bup.oval;
		;
		goto R999;

	case 49: // STATE 18
		;
		p_restor(II);
		;
		;
		goto R999;
	}

