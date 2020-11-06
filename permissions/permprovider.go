package permissions

import (
	"os"

	"lbauthdata/logger"
	"lbauthdata/model"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

type Db struct {
	conn *pgx.Conn
}

var log = logger.GetLogger("permissions")

func NewDBPermissionProvider(connstr string) (*Db, error) {
	log.Info("Creating database connection:", zap.String("dbconfig:", connstr))

	pgConfig, err := pgx.ParseConnectionString(connstr)
	if err != nil {
		log.Error("Error parsing DB connection string:", zap.String("error:", err.Error()))
		os.Exit(1)
	}

	// initialize db connection
	dbconn, err := pgx.Connect(pgConfig)
	if err != nil {
		log.Error("Error during database connection:", zap.String("error:", err.Error()))
		os.Exit(1)
	}

	return &Db{conn: dbconn}, nil
}

func (db *Db) GetGroupsPermissions(groupsarray []string, reqId string) (model.GroupPermMappings, error) {

	groupsquery := ""

	// Generating query string
	for index, group := range groupsarray {
		groupsquery += " group_uuid = '" + group + "'"
		if index < (len(groupsarray) - 1) {
			groupsquery += " OR"
		}
	}

	log.Info("groupsquery", zap.String("query:", groupsquery))

	// var id int64
	// var group_uuid pgtype.UUID
	// var role_uuid pgtype.UUID
	// Send the query to the server. The returned rows MUST be closed
	// before conn can be used again.
	rows, err := db.conn.Query(
		`SELECT
		roles_group_mapping.group_uuid,
		bool_or (roles.admin_iots) AS admin_iots,
		bool_or (roles.view_iots) AS view_iots,
		bool_or (roles.configure_iots) AS configure_iots,
		bool_or (roles.vpn_iots) AS vpn_iots,
		bool_or (roles.webpage_iots) AS webpage_iots,
		bool_or (roles.hmi_iots) AS hmi_iots,
		bool_or (roles.data_admin) AS data_admin,
		bool_or (roles.data_read) AS data_read,
		bool_or (roles.data_cold_read) AS data_cold_read,
		bool_or (roles.data_warm_read) AS data_warm_read,
		bool_or (roles.data_hot_read) AS data_hot_read,
		bool_or (roles.services_admin) AS services_admin,
		bool_or (roles.billing_admin) AS billing_admin,
		bool_or (roles.org_admin) AS org_admin
	FROM	roles_group_mapping
	INNER JOIN roles ON roles_group_mapping.role_uuid = roles.uuid AND (` + groupsquery + `) GROUP BY roles_group_mapping.group_uuid;`)
	if err != nil {
		log.Error("during db query:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
		panic(err)
	}
	// rows.Close is called by rows.Next when all rows are read
	// or an error occurs in Next or Scan. So it may optionally be
	// omitted if nothing in the rows.Next loop can panic. It is
	// safe to close rows multiple times.
	defer rows.Close()

	// var sum int32

	groupsArr := model.GroupPermMappings{}

	// Iterate through the result set
	for rows.Next() {

		groupMap := model.Mapping{}

		var group_uuid pgtype.UUID
		var admin_iots pgtype.Bool
		var view_iots pgtype.Bool
		var configure_iots pgtype.Bool
		var vpn_iots pgtype.Bool
		var webpage_iots pgtype.Bool
		var hmi_iots pgtype.Bool
		var data_admin pgtype.Bool
		var data_read pgtype.Bool
		var data_cold_read pgtype.Bool
		var data_warm_read pgtype.Bool
		var data_hot_read pgtype.Bool
		var services_admin pgtype.Bool
		var billing_admin pgtype.Bool
		var org_admin pgtype.Bool

		err = rows.Scan(
			&group_uuid,
			&admin_iots,
			&view_iots,
			&configure_iots,
			&vpn_iots,
			&webpage_iots,
			&hmi_iots,
			&data_admin,
			&data_read,
			&data_cold_read,
			&data_warm_read,
			&data_hot_read,
			&services_admin,
			&billing_admin,
			&org_admin)

		if err != nil {
			log.Error("error scanning rows:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
			panic(err)
		}

		group_uuid.AssignTo(&groupMap.Group_uuid)
		admin_iots.AssignTo(&groupMap.Permissions.Admin_iots)
		view_iots.AssignTo(&groupMap.Permissions.View_iots)
		configure_iots.AssignTo(&groupMap.Permissions.Configure_iots)
		vpn_iots.AssignTo(&groupMap.Permissions.Vpn_iots)
		webpage_iots.AssignTo(&groupMap.Permissions.Webpage_iots)
		hmi_iots.AssignTo(&groupMap.Permissions.Hmi_iots)
		data_admin.AssignTo(&groupMap.Permissions.Data_admin)
		data_read.AssignTo(&groupMap.Permissions.Data_read)
		data_cold_read.AssignTo(&groupMap.Permissions.Data_cold_read)
		data_warm_read.AssignTo(&groupMap.Permissions.Data_warm_read)
		data_hot_read.AssignTo(&groupMap.Permissions.Data_hot_read)
		services_admin.AssignTo(&groupMap.Permissions.Services_admin)
		billing_admin.AssignTo(&groupMap.Permissions.Billing_admin)
		org_admin.AssignTo(&groupMap.Permissions.Org_admin)

		groupsArr.Groups = append(groupsArr.Groups, groupMap)
		// sum += id
	}

	log.Info("permissions were retreived for groups", zap.String("reqid:", reqId))

	// Any errors encountered by rows.Next or rows.Scan will be returned here
	if rows.Err() != nil {
		log.Error("error during rows next:", zap.String("error:", rows.Err().Error()), zap.String("reqid:", reqId))
		panic(rows.Err())
	}

	return groupsArr, nil
}
