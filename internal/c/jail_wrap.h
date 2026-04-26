#ifndef GJ_JAIL_WRAP_H
#define GJ_JAIL_WRAP_H

#include <sys/types.h>

struct gj_param {
	const char *name;
	const char *value;
	int is_bool;
};

struct gj_result {
	int jid;
	int errnum;
	char errmsg[1024];
};

/*
 * gj_jail_set creates or updates a jail using jailparam_set.
 * flags: JAIL_CREATE (0x01), JAIL_UPDATE (0x02), or both (0x03).
 */
struct gj_result gj_jail_set(struct gj_param *params, int nparams, int flags);

/*
 * gj_jail_get retrieves jail parameters.
 * The first parameter should be the lookup key (name or jid with a value).
 * Remaining parameters are names to fetch (values filled in by export).
 * Returns jid on success, -1 on error.
 */
struct gj_result gj_jail_get(struct gj_param *params, int nparams, int flags);

/*
 * gj_jail_get_export exports a single parameter value after a successful gj_jail_get.
 * Caller must free the returned string with gj_free.
 * Returns NULL on error.
 */
char *gj_jail_get_export(const char *jid_str, const char *param_name);

/*
 * gj_jail_remove removes a jail by JID.
 * Returns result with jid=0 on success.
 */
struct gj_result gj_jail_remove(int jid);

/*
 * gj_jail_attach attaches the calling process to a jail.
 * Returns result with jid=0 on success.
 */
struct gj_result gj_jail_attach(int jid);

void gj_free(void *ptr);

#endif /* GJ_JAIL_WRAP_H */
