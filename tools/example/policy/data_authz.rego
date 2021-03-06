package authzdata 

import input

default allow = false                              # unless otherwise defined, allow is false

allow = true {                                     # allow is true if...
    count(data_allowed_groups) > 0                 # there are at least one group that can access data.
}

data_allowed_groups[group] {                 	   # a group is in the data_allowed set if...
    some group
    hot_allowed[group]                             # it exists in the 'hot_allowed' set and...
}

data_allowed_groups[group] {	                   # a group is in the data_allowed set if...
    some group
    warm_allowed[group]                            # it exists in the 'warm_allowed' set and...
}

data_allowed_groups[group] {                       # a group is in the data_allowed set if...
    some group
    cold_allowed[group]                            # it exists in the 'cold_allowed' set and...
}

data_allowed_groups[group] {                       # a group is in the data_allowed set if...
    some group
    read_allowed[group]                            # it exists in the 'read_allowed' set and...
}


hot_allowed[group.group_uuid] {                             			# a group exists in the hot_allowed set if...
    group := input.groups[_]                        			# it exists in the input.groups collection and...
    group.permissions.data_hot_read == true     		        # among its roles there is one with data_hot_read permission...   
}

warm_allowed[group.group_uuid] {                             			# a group exists in the warm_allowed set if...
    group := input.groups[_]                        			# it exists in the input.groups collection and...
    group.permissions.data_warm_read == true     		        # among its roles there is one with data_warm_read permission...   
}

cold_allowed[group.group_uuid] {                             			# a group exists in the cold_allowed set if...
    group := input.groups[_]                        			# it exists in the input.groups collection and...
    group.permissions.data_cold_read == true     		        # among its roles there is one with data_cold_read permission...   
}

read_allowed[group.group_uuid] {                             			# a group exists in the read_allowed set if...
    group := input.groups[_]                        			# it exists in the input.groups collection and...
    group.permissions.data_read == true             			# among its roles there is one with data_read permission...   
}

