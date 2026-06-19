package customlayout

import (
	"log/slog"
	"net/http"

	"github.com/downballot/downballot/downballotapi"
	"github.com/downballot/downballot/iam"
	"github.com/downballot/downballot/permissionset"
	"github.com/downballot/ui/api"
	"github.com/downballot/ui/component/layout"
	"github.com/downballot/ui/component/page"
	"github.com/go-app-blazar/router"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const (
	MetaAutocrumbs   = "autocrumbs"
	MetaRequireLogin = "require-login"
	MetaPermission   = "permission"
	MetaTitle        = "title"
)

var Routes = []router.Route{
	{
		Path: "/",
		Component: func() app.Composer {
			return &layout.CenterLayout{}
		},
		Meta: map[string]string{
			MetaRequireLogin: "false",
		},
		Children: []router.Route{
			{
				Path: "/login",
				Component: func() app.Composer {
					return &page.LoginPage{}
				},
			},
			{
				Path: "/signup",
				Component: func() app.Composer {
					return &page.SignupPage{}
				},
			},
		},
	},
	{
		Path: "/",
		Component: func() app.Composer {
			return &DownballotLayout{}
		},
		Meta: map[string]string{
			MetaRequireLogin: "true",
		},
		Children: []router.Route{
			{
				Path: "/",
				Component: func() app.Composer {
					return &page.HomePage{}
				},
				Meta: map[string]string{
					MetaTitle: "Home",
				},
			},
			{
				Path:      "/organization",
				Component: nil,
				Children: []router.Route{
					{
						Path: "/",
						Component: func() app.Composer {
							return &page.OrganizationPage{}
						},
						Meta: map[string]string{
							MetaTitle: "Organizations",
						},
					},
					{
						Path: "/:organization_id",
						PathVariables: func(ctx app.Context, variables map[string]string) {
							{
								var output downballotapi.GetOrganizationResponse
								err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"], nil, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not get organization", "err", err)
									return
								}

								variables["organization_name"] = output.Organization.Name
							}

							{
								var output downballotapi.ListPermissionsResponse
								err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"]+"/permission", nil, &output)
								if err != nil {
									slog.ErrorContext(ctx.Context, "Could not get permissions", "err", err)
									return
								}

								var permissionSet permissionset.PermissionSet
								for _, permissionString := range output.Permissions {
									permissionSet.AddPermission(permissionset.Permission(permissionString))
								}

								ctx.SetState("organization/"+variables["organization_id"]+"/permission-set", permissionSet)
							}
						},
						Meta: map[string]string{
							MetaAutocrumbs: "true",
						},
						Component: func() app.Composer {
							return &OrganizationLayout{}
						},
						Children: []router.Route{
							{
								Path: "/",
								Component: func() app.Composer {
									return &page.OrganizationIDPage{}
								},
								Meta: map[string]string{
									MetaTitle: ":organization_name",
								},
							},
							{
								Path:      "/filter",
								Component: nil,
								Meta: map[string]string{
									MetaPermission: string(iam.IAMFilterRead),
								},
								Children: []router.Route{
									{
										Path: "/",
										Component: func() app.Composer {
											return &page.OrganizationIDFilterPage{}
										},
										Meta: map[string]string{
											MetaTitle: "Filters",
										},
									},
									{
										Path: "/new",
										Component: func() app.Composer {
											return &page.OrganizationIDFilterNewPage{}
										},
										Meta: map[string]string{
											MetaTitle: "New Filter",
										},
									},
									{
										Path:      "/:filter_id",
										Component: nil,
										PathVariables: func(ctx app.Context, variables map[string]string) {
											var output downballotapi.GetFilterResponse
											err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"]+"/filter/"+variables["filter_id"], nil, &output)
											if err != nil {
												slog.ErrorContext(ctx.Context, "Could not get filter", "err", err)
												return
											}

											variables["filter_name"] = output.Filter.Name
										},
										Children: []router.Route{
											{
												Path: "/",
												Component: func() app.Composer {
													return &page.OrganizationIDFilterIDEditPage{} // TODO: CHANGE THIS
												},
												Meta: map[string]string{
													MetaTitle: ":filter_name",
												},
											},
											{
												Path: "/edit",
												Component: func() app.Composer {
													return &page.OrganizationIDFilterIDEditPage{}
												},
												Meta: map[string]string{
													MetaTitle: "Edit Filter",
												},
											},
										},
									},
								},
							},
							{
								Path:      "/group",
								Component: nil,
								Meta: map[string]string{
									MetaPermission: string(iam.IAMGroupRead),
								},
								Children: []router.Route{
									{
										Path: "/",
										Component: func() app.Composer {
											return &page.OrganizationIDGroupPage{}
										},
										Meta: map[string]string{
											MetaTitle: "Groups",
										},
									},
									{
										Path: "/new",
										Component: func() app.Composer {
											return &page.OrganizationIDGroupNewPage{}
										},
										Meta: map[string]string{
											MetaTitle: "New Group",
										},
									},
									{
										Path:      "/:group_id",
										Component: nil,
										PathVariables: func(ctx app.Context, variables map[string]string) {
											var output downballotapi.GetGroupResponse
											err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"]+"/group/"+variables["group_id"], nil, &output)
											if err != nil {
												slog.ErrorContext(ctx.Context, "Could not get group", "err", err)
												return
											}

											variables["group_name"] = output.Group.Name
										},
										Children: []router.Route{
											{
												Path: "/",
												Component: func() app.Composer {
													return &page.OrganizationIDGroupIDPage{}
												},
												Meta: map[string]string{
													MetaTitle: ":group_name",
												},
											},
											{
												Path: "/edit",
												Component: func() app.Composer {
													return &page.OrganizationIDGroupIDEditPage{}
												},
												Meta: map[string]string{
													MetaTitle: "Edit Group",
												},
											},
											{
												Path: "/person",
												Component: func() app.Composer {
													return &page.OrganizationIDGroupIDPersonPage{}
												},
												Meta: map[string]string{
													MetaTitle: "Persons",
												},
											},
											{
												Path: "/person-mailing-labels",
												Component: func() app.Composer {
													return &page.OrganizationIDGroupIDPersonMailingLabelsPage{}
												},
												Meta: map[string]string{
													MetaTitle: "Mailing Labels",
												},
											},
										},
									},
								},
							},
							{
								Path:      "/person",
								Component: nil,
								Meta: map[string]string{
									MetaPermission: string(iam.IAMPersonRead),
								},
								Children: []router.Route{
									{
										Path:      "/:voter_id",
										Component: nil,
										PathVariables: func(ctx app.Context, variables map[string]string) {
											var output downballotapi.GetPersonResponse
											err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"]+"/person/"+variables["voter_id"], nil, &output)
											if err != nil {
												slog.ErrorContext(ctx.Context, "Could not get person", "err", err)
												return
											}

											for name, value := range output.Person.Fields {
												variables["person_field_"+name] = value
											}
										},
										Children: []router.Route{
											{
												Path: "/",
												Component: func() app.Composer {
													return &page.OrganizationIDPersonIDPage{}
												},
												Meta: map[string]string{
													MetaTitle: ":person_field_name",
												},
											},
										},
									},
								},
							},
							{
								Path:      "/person-field",
								Component: nil,
								Meta: map[string]string{
									MetaPermission: string(iam.IAMPersonFieldDefinitionRead),
								},
								Children: []router.Route{
									{
										Path: "/",
										Component: func() app.Composer {
											return &page.OrganizationIDPersonFieldPage{}
										},
										Meta: map[string]string{
											MetaTitle: "Person Fields",
										},
									},
									{
										Path: "/new",
										Component: func() app.Composer {
											return &page.OrganizationIDPersonFieldNewPage{}
										},
										Meta: map[string]string{
											MetaTitle: "New Person Field",
										},
									},
									{
										Path:      "/:person_field_id",
										Component: nil,
										PathVariables: func(ctx app.Context, variables map[string]string) {
											var output downballotapi.GetPersonFieldResponse
											err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"]+"/person-field/"+variables["person_field_id"], nil, &output)
											if err != nil {
												slog.ErrorContext(ctx.Context, "Could not get person field", "err", err)
												return
											}
											variables["person_field_name"] = output.PersonField.Name
										},
										Children: []router.Route{
											{
												Path: "/",
												Component: func() app.Composer {
													return &page.OrganizationIDPersonFieldIDPage{}
												},
												Meta: map[string]string{
													MetaTitle: ":person_field_name",
												},
											},
											{
												Path: "/edit",
												Component: func() app.Composer {
													return &page.OrganizationIDPersonFieldIDEditPage{}
												},
												Meta: map[string]string{
													MetaTitle: "Edit Person Field",
												},
											},
										},
									},
								},
							},
							{
								Path:      "/user",
								Component: nil,
								Meta: map[string]string{
									MetaPermission: string(iam.IAMOrganizationUserRead),
								},
								Children: []router.Route{
									{
										Path: "/",
										Component: func() app.Composer {
											return &page.OrganizationIDUserPage{}
										},
										Meta: map[string]string{
											MetaTitle: "Users",
										},
									},
									{
										Path: "/new",
										Component: func() app.Composer {
											return &page.OrganizationIDUserNewPage{}
										},
										Meta: map[string]string{
											MetaTitle: "Add User",
										},
									},
									{
										Path:      "/:user_id",
										Component: nil,
										PathVariables: func(ctx app.Context, variables map[string]string) {
											var output downballotapi.GetUserResponse
											err := api.Do(ctx, http.MethodGet, "/api/v1/organization/"+variables["organization_id"]+"/user/"+variables["user_id"], nil, &output)
											if err != nil {
												slog.ErrorContext(ctx.Context, "Could not get user", "err", err)
												return
											}
											variables["user_name"] = output.User.Name
										},
										Children: []router.Route{
											{
												Path: "/",
												Component: func() app.Composer {
													return &page.OrganizationIDUserIDPage{}
												},
												Meta: map[string]string{
													MetaTitle: ":user_name",
												},
											},
											{
												Path: "/edit",
												Component: func() app.Composer {
													return &page.OrganizationIDUserIDEditPage{}
												},
												Meta: map[string]string{
													MetaTitle: "Edit User",
												},
											},
											{
												Path:      "/group",
												Component: nil,
												Children: []router.Route{
													{
														Path:      "/:group_id",
														Component: nil,
														Children: []router.Route{
															{
																Path: "/edit",
																Component: func() app.Composer {
																	return &page.OrganizationIDUserIDGroupIDEditPage{}
																},
																Meta: map[string]string{
																	MetaTitle: "Edit User Group",
																},
															},
														},
													},
													{
														Path: "/new",
														Component: func() app.Composer {
															return &page.OrganizationIDUserIDGroupNewPage{}
														},
														Meta: map[string]string{
															MetaTitle: "Add User To Group",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Path: "/profile",
				Component: func() app.Composer {
					return &page.ProfilePage{}
				},
				Meta: map[string]string{
					MetaTitle: "Profile",
				},
			},
		},
	},
}
