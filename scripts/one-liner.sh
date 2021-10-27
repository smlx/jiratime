# This function is an example to show how jiratime can be used imperatively.
# It is discussed in https://github.com/smlx/jiratime/issues/19.
#
# Example usage:
#
# $ jt 0900-0915 DAF-7 test some time
jt()
{
	local t=$1 i=$2
	shift 2
	printf "%s\n%s\n%s\n" "$t" "$i" "$*" | jiratime
}
