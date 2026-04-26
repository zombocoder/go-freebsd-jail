#include "jail_wrap.h"

#include <sys/param.h>
#include <sys/jail.h>
#include <errno.h>
#include <jail.h>
#include <stdlib.h>
#include <string.h>

/* jail_errmsg is a global in libjail — not thread-safe.
 * The Go layer must serialize calls with a mutex. */
extern char jail_errmsg[];

static void result_ok(struct gj_result *res, int jid)
{
	res->jid = jid;
	res->errnum = 0;
	res->errmsg[0] = '\0';
}

static void result_err(struct gj_result *res)
{
	res->jid = -1;
	res->errnum = errno;
	if (jail_errmsg[0] != '\0') {
		strlcpy(res->errmsg, jail_errmsg, sizeof(res->errmsg));
		jail_errmsg[0] = '\0';
	} else {
		strlcpy(res->errmsg, strerror(errno), sizeof(res->errmsg));
	}
}

struct gj_result
gj_jail_set(struct gj_param *params, int nparams, int flags)
{
	struct gj_result res;
	struct jailparam *jp;
	int i, jid;

	jp = calloc(nparams, sizeof(struct jailparam));
	if (jp == NULL) {
		res.jid = -1;
		res.errnum = ENOMEM;
		strlcpy(res.errmsg, "calloc failed", sizeof(res.errmsg));
		return res;
	}

	for (i = 0; i < nparams; i++) {
		if (jailparam_init(&jp[i], params[i].name) != 0) {
			result_err(&res);
			jailparam_free(jp, i);
			free(jp);
			return res;
		}
		if (!params[i].is_bool && params[i].value != NULL) {
			if (jailparam_import(&jp[i], params[i].value) != 0) {
				result_err(&res);
				jailparam_free(jp, i + 1);
				free(jp);
				return res;
			}
		}
	}

	jid = jailparam_set(jp, nparams, flags);
	if (jid < 0) {
		result_err(&res);
	} else {
		result_ok(&res, jid);
	}

	jailparam_free(jp, nparams);
	free(jp);
	return res;
}

struct gj_result
gj_jail_get(struct gj_param *params, int nparams, int flags)
{
	struct gj_result res;
	struct jailparam *jp;
	int i, jid;

	jp = calloc(nparams, sizeof(struct jailparam));
	if (jp == NULL) {
		res.jid = -1;
		res.errnum = ENOMEM;
		strlcpy(res.errmsg, "calloc failed", sizeof(res.errmsg));
		return res;
	}

	for (i = 0; i < nparams; i++) {
		if (jailparam_init(&jp[i], params[i].name) != 0) {
			result_err(&res);
			jailparam_free(jp, i);
			free(jp);
			return res;
		}
		if (params[i].value != NULL && params[i].value[0] != '\0') {
			if (jailparam_import(&jp[i], params[i].value) != 0) {
				result_err(&res);
				jailparam_free(jp, i + 1);
				free(jp);
				return res;
			}
		}
	}

	jid = jailparam_get(jp, nparams, flags);
	if (jid < 0) {
		result_err(&res);
		jailparam_free(jp, nparams);
		free(jp);
		return res;
	}

	result_ok(&res, jid);

	/* Export values back into params */
	for (i = 0; i < nparams; i++) {
		char *val = jailparam_export(&jp[i]);
		if (val != NULL) {
			/* Caller is responsible for freeing via gj_free */
			params[i].value = val;
		}
	}

	jailparam_free(jp, nparams);
	free(jp);
	return res;
}

char *
gj_jail_get_export(const char *jid_str, const char *param_name)
{
	struct jailparam jp[2];
	int jid;
	char *val;

	if (jailparam_init(&jp[0], "jid") != 0)
		return NULL;
	if (jailparam_import(&jp[0], jid_str) != 0) {
		jailparam_free(jp, 1);
		return NULL;
	}
	if (jailparam_init(&jp[1], param_name) != 0) {
		jailparam_free(jp, 1);
		return NULL;
	}

	jid = jailparam_get(jp, 2, 0);
	if (jid < 0) {
		jailparam_free(jp, 2);
		return NULL;
	}

	val = jailparam_export(&jp[1]);
	jailparam_free(jp, 2);
	return val; /* caller frees with gj_free */
}

struct gj_result
gj_jail_remove(int jid)
{
	struct gj_result res;
	if (jail_remove(jid) < 0) {
		result_err(&res);
	} else {
		result_ok(&res, 0);
	}
	return res;
}

struct gj_result
gj_jail_attach(int jid)
{
	struct gj_result res;
	if (jail_attach(jid) < 0) {
		result_err(&res);
	} else {
		result_ok(&res, 0);
	}
	return res;
}

void
gj_free(void *ptr)
{
	free(ptr);
}
