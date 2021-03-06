package ilang

import "strconv"
import "strings"
//import "github.com/gedex/inflector"

func (ic *Compiler) expression() string {
	var token = ic.Scan(0)
	
	for token == "\n" {
		token = ic.Scan(0)
	}
	
	switch token {
		case "true":
			ic.ExpressionType = Number
			return "1"
		case "false":
			ic.ExpressionType = Number
			return "0"
		case "error":
			ic.ExpressionType = Number
			return "ERROR"
	}
	
	//Text.
	if token[0] == '"' || token[0] == '`' {
		ic.ExpressionType = Text
		return ic.ParseString(token)
	}
	
	//Hexadecimal.
	if len(token) > 2 && token[0] == '0' && token[1] == 'x' { 
		ic.ExpressionType = Number
		return token
	}
	
	//Subexpessions.
	if token == "(" {
		defer func() {
			ic.Scan(')')
		}()
		return ic.ScanExpression()
	}
	
	//Is it a literal number? Then just return it.
	if _, err := strconv.Atoi(token); err == nil{
		ic.ExpressionType = Number
		return token
	}
	
	//Minus.
	if token == "-" {
		ic.NextToken = token
		ic.ExpressionType = Number
		return ic.Shunt("0")
	}
	
	if t := ic.GetVariable(token); t != Undefined {
	
		if ic.TakingExpression && ic.GetVariable(token+".") == Protected && ic.Peek() != "." {
			ic.RaiseError("Cannot transfer ownership of protected variable ", token)
		}
		
		ic.ExpressionType = t
		ic.SetVariable(token+"_use", Used)
		
		if ic.Peek() == "." {
			return ic.Shunt(token)
		}
		
		return token
	}
	
	if ic.TypeExists(token) {
		ic.ExpressionType = ic.DefinedTypes[token]
		
		//This is a constructor. eg. var bug = Bug(); where Bug is a type.
		if ic.Peek() == "(" || ic.NextToken == "(" {
			ic.Scan('(')
			ic.Scan(')')
			
			return ic.CallType(token)
			
		//This is a type literal.
		} else if ic.Peek() == "{" {
			ic.NextToken = token
			variable := ic.ScanTypeLiteral()
				//TODO better gc protection.
			ic.SetVariable(variable, ic.DefinedTypes[token])
			ic.SetVariable(variable+"_use", Used)
			return variable
			
		}
	}
	
	if token == "new" {
		ic.Scan('(')
		var sort = ic.DefinedTypes[ic.Scan(Name)]
		ic.Scan(')') 
		if _, ok := ic.DefinedFunctions["new_m_"+sort.Name]; !ok {
			ic.RaiseError("no new method found for ", sort.Name)
		}
		var r = ic.Tmp("new")
		ic.Assembly("RUN new_m_"+sort.Name)
		ic.Assembly("GRAB ", r)
		ic.ExpressionType = sort
		return r
	}
	
	for _, expression := range Expressions {
		id := expression(ic)
		if id != "" {
			return id
		}
	}
	
	if ic.Translation && !ic.Translated {
		ic.Translated = true
		defer func() {
			ic.Translated = false
		}()
		var err error
		ic.NextToken, err = getTranslation(ic.Language, "en", token)
		if strings.Contains(ic.NextToken, "\n") {
			ic.NextToken = strings.Split(ic.NextToken, "\n")[0]
		}
		println(ic.NextToken, ic.Language)
		if err != nil {
			ic.RaiseError(err)
		}
		return ic.expression() 
	}
	
	/*token = inflector.Singularize(token)
	if t, ok := ic.DefinedTypes[token]; ok {
		ic.Scan('(')
		ic.Scan(')')
		return ic.NewListOf(t)
	}
	if t, ok := ic.DefinedInterfaces[token]; ok {
		ic.Scan('(')
		ic.Scan(')')
		return ic.NewListOf(t.GetType())
	}*/
	
	ic.NextToken = token
	ic.ExpressionType = Undefined
	//ic.RaiseError()
	return ""
}

func (ic *Compiler) ScanExpression() string {
	return ic.Shunt(ic.expression())
}
