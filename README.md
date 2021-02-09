# **MTAUTHDATA** 
MetricTank ABAC authorization proxy enforcer for graphite metric queries

## Architecture 
Mtauthdata is composed of the following modules
- lbDataAuthzProxy: Http proxy module implemented by lightweight go-chi module 
- PermissionProvider: DInjectable permission provider gathering user permissions and attributes
- AuthzProvider: DInjectable authorization provider enforcing policy decisions 

*Proxy middleware chaining:* 
```
         Httpreq -> GroupPermissionsMiddleware -> groupmappings
                                  │
                                  v
      groupmappings -> AuthzEnforcementMiddleware -> grouptemps
                                  │
                                  ├─────────────────────────────────────────────────┐
                                  │                                                 │ 
                                  v                                                 v
  grouptemps -> TagsFilteringMiddleware -> rawquery         grouptemps -> RenderFilteringMiddleware -> rawquery
                                  │                                                 │                                             
                                  └─────────────────────────────────────────────────┘
                                                       │ 
                                                       v
                                      rawquery -> proxyhandler
```

## **Example req** 

*Look for all visible tag keys with prefix*
```
Req:
curl -H "X-Org-Id: 2" "http://localhost:9001/tags/autoComplete/tags?tagPrefix=na"
```

*Look for all visible tag keys* 

```
Req:
curl -H "X-Org-Id: 2" "http://localhost:9001/render?target=seriesByTag('name=~(^demotags.iot1.metric0$)','data:pr:ext:acl:grouptemp=~(^group:ou:e34ba21c74c289ba894b75ae6c76d22f:temp:hot$)')&from=-5min&until=now&format=json&maxDataPoints=653
```
