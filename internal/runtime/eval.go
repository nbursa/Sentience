package runtime

import (
	"fmt"
	"io"
	"strings"

	"github.com/nbursa/sentience/internal/types"
)

var outWriter io.Writer = nil

func printOut(format string, args ...any) {
	if outWriter != nil {
		fmt.Fprintf(outWriter, format, args...)
	} else {
		fmt.Printf(format, args...)
	}
}

func Eval(node types.Node, indent string, ctx *AgentContext) {
	switch n := node.(type) {

	case *types.Program:
		for _, stmt := range n.Statements {
			Eval(stmt, indent, ctx)
		}

	case *types.AgentStatement:
		printOut("%sAgent: %s\n", indent, n.Name)
		ctx.CurrentAgent = n
		for _, stmt := range n.Body {
			if _, isTrain := stmt.(*types.TrainStatement); !isTrain {
				Eval(stmt, indent+"  ", ctx)
			}
		}
		printOut("%sAgent: %s [registered]\n", indent, n.Name)

	case *types.MemStatement:
		printOut("%sInit mem: %s\n", indent, n.Target)
		ctx.SetMem(n.Target, "__init__", "1")

	case *types.OnInputStatement:
		printOut("%sOn Input: (%s)\n", indent, n.Param)
		for _, stmt := range n.Body {
			Eval(stmt, indent+"  ", ctx)
		}

	case *types.ReflectStatement:
		printOut("%sReflect block:\n", indent)
		for _, stmt := range n.Body {
			Eval(stmt, indent+"  ", ctx)
		}

	case *types.TrainStatement:
		printOut("%sTrain block:\n", indent)
		for _, stmt := range n.Body {
			Eval(stmt, indent+"  ", ctx)
		}

	case *types.GoalStatement:
		printOut("%sGoal: \"%s\"\n", indent, n.Value)

	case *types.EmbedStatement:
		printOut("%sEmbed: %s -> %s\n", indent, n.Source, n.Target)

		val := ctx.GetMem("short", n.Source)

		var target string
		if n.Target == "mem.short" {
			target = "short"
		} else if n.Target == "mem.long" {
			target = "long"
		} else {
			target = "short"
		}
		ctx.SetMem(target, n.Source, val)

	case *types.LinkStatement:
		printOut("%sLink: %s <-> %s\n", indent, n.From, n.To)
		ctx.Links[n.From] = n.To
		ctx.Links[n.To] = n.From

	case *types.IfStatement:
		printOut("%sIf: %s\n", indent, n.Condition)

		if strings.Contains(n.Condition, "loss") {
			printOut("%s  [stub eval: assuming loss > 0.1]\n", indent)
			for _, stmt := range n.Body {
				Eval(stmt, indent+"  ", ctx)
			}
			return
		}

		if strings.HasPrefix(n.Condition, "context includes ") {
			key := strings.TrimPrefix(n.Condition, "context includes ")
			key = strings.Trim(key, `"' `)

			found := false
			for k := range ctx.MemShort {
				if strings.Contains(k, key) || strings.Contains(ctx.MemShort[k], key) {
					found = true
					break
				}
			}

			if found {
				for _, stmt := range n.Body {
					Eval(stmt, indent+"  ", ctx)
				}
			} else {
				printOut("%s  [context miss: \"%s\"]\n", indent, key)
			}
		} else {
			printOut("%sCondition not supported: %s\n", indent, n.Condition)
		}

	case *types.EnterStatement:
		printOut("%sEnter: %s\n", indent, n.Target)

	case *types.ReflectAccessStatement:
		// fmt.Println("[debug] executing ReflectAccessStatement")
		val := ctx.GetMem(n.MemTarget, n.Key)
		printOut("%smem.%s[\"%s\"] = \"%s\"\n", indent, n.MemTarget, n.Key, val)

	case *types.PrintStatement:
		printOut("%s%s\n", indent, n.Value)

	case *types.EvolveStatement:
		printOut("%sEvolve block:\n", indent)
		for _, stmt := range n.Body {
			Eval(stmt, indent+"  ", ctx)
		}

	case *types.ReflectLatentStatement:
		results := ctx.SimilarTo(n.Query)
		printOut("%smem.latent similar_to(\"%s\") → %v\n", indent, n.Query, results)

	default:
		printOut("%sUnknown node: %T\n", indent, n)
	}
}
