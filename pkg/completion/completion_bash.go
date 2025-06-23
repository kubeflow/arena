// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package completion

var (
	BashCompletionFlags = map[string]string{
		"namespace": "__arena_get_namespace",
	}
)

const (
	BashCompletionFunc = "__arena_parse_get()\n{\n\tlocal arena_output out\n\tif arena_output=$(arena list $(__arena_override_flags) | grep -v -E 'NAME.*STATUS.*TRAINER.*AGE' 2>/dev/null); then\n\t\tout=($(echo \"${arena_output}\" | awk '{print $1}'))\n\t\tCOMPREPLY=( $( compgen -W \"${out[*]}\" -- \"$cur\" )\n\tfi\n}\n\n__arena_parse_serve_get() {\n\tlocal arena_output out\n\tif arena_output=$(arena serve list $(__arena_override_flags) | grep -v -E 'NAME.*TYPE.*VERSION' 2>/dev/null); then\n\t\tout=($(echo \"${arena_output}\" | awk '{print $1}'))\n\t\tCOMPREPLY=( $( compgen -W \"${out[*]}\" -- \"$cur\" )\n\tfi\n}\n__arena_serve_all_namespace() {\n\tlocal arena_output out\n\tif arena_output=$(arena serve list --all-namespaces | grep -v -E 'NAME.*TYPE.*NAMESPACE' 2>/dev/null); then\n\t\tout=($(echo \"${arena_output}\" | awk '{print $3}'))\n\t\tCOMPREPLY=( $( compgen -W \"${out[*]}\" -- \"$cur\" )\n\n\tfi\n}\n\n__arena_serve_all_version() {\n\tlocal arena_output out\n\tif arena_output=$(arena serve list $(__arena_override_flags) | grep -v -E 'NAME.*TYPE.*VERSION' 2>/dev/null); then\n\t\tout=($(echo \"${arena_output}\" | awk '{print $3}'))\n\t\tCOMPREPLY=( $( compgen -W \"${out[*]}\" -- \"$cur\" )\n\tfi\n}\n__arena_serve_all_type() {\n\tlocal arena_output out\n\tout=(tf trt custom)\n\tCOMPREPLY=( $( compgen -W \"${out[*]}\" -- \"$cur\" )\n}\n\n__custom_func() {\n    case ${last_command} in\n        arena_get | arena_logs | arena_delete | arena_logviewer | arena_top_job)\n            __arena_parse_get\n            return\n            ;;\n        arena_serve_get | arena_serve_logs | arena_serve_delete)\n            __arena_parse_serve_get\n            return\n            ;;\n        *)\n            ;;\n    esac\n}\n\n__arena_override_flag_list=(--namespace=)\n__arena_override_flags()\n{\n    local ${__arena_override_flag_list[*]##*-} two_word_of of var\n    for w in \"${words[@]}\"; do\n        if [ -n \"${two_word_of}\" ]; then\n            eval \"${two_word_of##*-}=\\\"${two_word_of}=\\${w}\\\"\"\n            two_word_of=\n            continue\n        fi\n        for of in \"${__arena_override_flag_list[@]}\"; do\n            case \"${w}\" in\n                ${of}=*)\n                    eval \"${of##*-}=\\\"${w}\\\"\"\n                    ;;\n                ${of})\n                    two_word_of=\"${of}\"\n                    ;;\n            esac\n        done\n    for var in \"${__arena_override_flag_list[@]##*-}\"; do\n        if eval \"test -n \\\"\\$${var}\\\"\"; then\n            eval \"echo \\${${var}}\"\n        fi\n    done"
)
