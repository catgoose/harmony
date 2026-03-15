package schema

// setup:feature:session_settings:start

var SessionSettingsTable = NewTable("SessionSettings").
	Columns(
		AutoIncrCol("Id"),
		Col("SessionUUID", TypeVarchar(36)).NotNull().Unique(),
		Col("Theme", TypeString(50)).NotNull().Default("'light'"),
	).
	WithTimestamps()

// setup:feature:session_settings:end

var ErrorTracesTable = NewTable("ErrorTraces").
	Columns(
		AutoIncrCol("Id"),
		Col("RequestID", TypeVarchar(64)).NotNull().Unique(),
		Col("ErrorChain", TypeText()).NotNull(),
		Col("StatusCode", TypeInt()).NotNull(),
		Col("Route", TypeString(500)).NotNull(),
		Col("Method", TypeString(10)).NotNull(),
		Col("UserAgent", TypeText()),
		Col("RemoteIP", TypeString(45)),
		Col("UserID", TypeString(255)),
		Col("Entries", TypeText()).NotNull(),
	).
	WithTimestamps().
	Indexes(
		Index("idx_error_traces_request_id", "RequestID"),
		Index("idx_error_traces_created_at", "CreatedAt"),
		Index("idx_error_traces_user_id", "UserID"),
	)

var UsersTable = NewTable("Users").
	Columns(
		AutoIncrCol("ID"),
		Col("AzureId", TypeVarchar(255)).NotNull().Unique(),
		Col("GivenName", TypeString(255)),
		Col("Surname", TypeString(255)),
		Col("DisplayName", TypeString(255)),
		Col("UserPrincipalName", TypeString(255)).NotNull(),
		Col("Mail", TypeString(255)),
		Col("JobTitle", TypeString(255)),
		Col("OfficeLocation", TypeString(255)),
		Col("Department", TypeString(255)),
		Col("CompanyName", TypeString(255)),
		Col("AccountName", TypeString(255)),
		Col("LastLoginAt", TypeTimestamp()),
	).
	WithTimestamps().
	Indexes(
		Index("idx_users_azureid", "AzureId"),
		Index("idx_users_userprincipalname", "UserPrincipalName"),
		Index("idx_users_displayname", "DisplayName"),
		Index("idx_users_mail", "Mail"),
		Index("idx_users_lastloginat", "LastLoginAt"),
	)
