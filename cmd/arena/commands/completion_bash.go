package commands

var (
	bash_completion_flags = map[string]string{
		"namespace": "__arena_get_namespace",
	}
)

const (
	bashCompletionFunc = `__arena_parse_get()
{
	local arena_output out
	if arena_output=$(arena list $(__arena_override_flags) | grep -v -E 'NAME.*STATUS.*TRAINER.*AGE' 2>/dev/null); then
		out=($(echo "${arena_output}" | awk '{print $1}'))
		COMPREPLY=( $( compgen -W "${out[*]}" -- "$cur" ) )
	fi
}


__custom_func() {
    case ${last_command} in
        arena_get | arena_logs | arena_delete | arena_logviewer )
            __arena_parse_get
            return
            ;;
        *)
            ;;
    esac
}

__arena_override_flag_list=(--namespace=)
__arena_override_flags()
{
    local ${__arena_override_flag_list[*]##*-} two_word_of of var
    for w in "${words[@]}"; do
        if [ -n "${two_word_of}" ]; then
            eval "${two_word_of##*-}=\"${two_word_of}=\${w}\""
            two_word_of=
            continue
        fi
        for of in "${__arena_override_flag_list[@]}"; do
            case "${w}" in
                ${of}=*)
                    eval "${of##*-}=\"${w}\""
                    ;;
                ${of})
                    two_word_of="${of}"
                    ;;
            esac
        done
    done
    for var in "${__arena_override_flag_list[@]##*-}"; do
        if eval "test -n \"\$${var}\""; then
            eval "echo \${${var}}"
        fi
    done
}
`
)
