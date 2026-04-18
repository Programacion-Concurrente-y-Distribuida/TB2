	switch (t->back) {
	default: Uerror("bad return move");
	case  0: goto R999; /* nothing to undo */

		 /* CLAIM sem_respetado */
;
		
	case 3: // STATE 1
		goto R999;

	case 4: // STATE 10
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* CLAIM progreso_wg */
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

		 /* CLAIM terminacion */
;
		;
		
	case 9: // STATE 6
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* CLAIM no_exceso */
;
		
	case 10: // STATE 1
		goto R999;

	case 11: // STATE 10
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* PROC :init: */
;
		;
		
	case 13: // STATE 2
		;
		_m = unsend(now.sem);
		;
		goto R999;

	case 14: // STATE 3
		;
		now.active_downloads = trpt->bup.oval;
		;
		goto R999;

	case 15: // STATE 5
		;
		;
		delproc(0, now._nr_pr-1);
		;
		goto R999;

	case 16: // STATE 6
		;
		((P1 *)_this)->i = trpt->bup.oval;
		;
		goto R999;
;
		;
		;
		;
		;
		;
		
	case 20: // STATE 15
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* PROC descargador */

	case 21: // STATE 1
		;
		((P0 *)_this)->exito = trpt->bup.oval;
		;
		goto R999;

	case 22: // STATE 2
		;
		((P0 *)_this)->exito = trpt->bup.oval;
		;
		goto R999;

	case 23: // STATE 5
		;
	/* 0 */	((P0 *)_this)->exito = trpt->bup.oval;
		;
		;
		goto R999;

	case 24: // STATE 6
		;
		_m = unsend(now.mu);
		;
		goto R999;
;
		;
		
	case 26: // STATE 8
		;
		now.mutex_ocupado = trpt->bup.oval;
		;
		goto R999;
;
		;
		
	case 28: // STATE 10
		;
		now.downloaded_count = trpt->bup.oval;
		;
		goto R999;

	case 29: // STATE 11
		;
		now.mutex_ocupado = trpt->bup.oval;
		;
		goto R999;

	case 30: // STATE 12
		;
		XX = 1;
		unrecv(now.mu, XX-1, 0, 1, 1);
		;
		;
		goto R999;

	case 31: // STATE 18
		;
		now.active_downloads = trpt->bup.ovals[1];
		now.finished_count = trpt->bup.ovals[0];
		;
		ungrab_ints(trpt->bup.ovals, 2);
		goto R999;

	case 32: // STATE 20
		;
		XX = 1;
		unrecv(now.sem, XX-1, 0, 1, 1);
		;
		;
		goto R999;

	case 33: // STATE 21
		;
		p_restor(II);
		;
		;
		goto R999;
	}

