package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/logger"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// Reflector implements self-reflection and verification for agent answers
// It can verify facts, calculations, and consistency before returning answers
type Reflector struct {
	provider types.LLMProvider
	memory   types.Memory
	registry *tools.Registry
	logger   logger.Logger
	verbose  bool
}

// NewReflector creates a new Reflector instance
func NewReflector(provider types.LLMProvider, memory types.Memory) *Reflector {
	return &Reflector{
		provider: provider,
		memory:   memory,
		registry: tools.NewRegistry(),
		logger:   logger.NewConsoleLogger(),
		verbose:  true,
	}
}

// WithTools adds tools to the reflector for verification
func (r *Reflector) WithTools(toolList ...tools.Tool) *Reflector {
	for _, tool := range toolList {
		r.registry.Register(tool)
	}
	return r
}

// WithLogger sets a custom logger
func (r *Reflector) WithLogger(log logger.Logger) *Reflector {
	r.logger = log
	return r
}

// SetVerbose controls whether reflection steps are logged
func (r *Reflector) SetVerbose(verbose bool) {
	r.verbose = verbose
}

// Reflect performs self-reflection on an answer to verify its correctness
// Returns a ReflectionCheck with verification results and potentially corrected answer
func (r *Reflector) Reflect(ctx context.Context, question string, initialAnswer string) (*types.ReflectionCheck, error) {
	if r.verbose {
		r.logger.Info("🔍 Starting self-reflection on answer...")
	}

	reflection := &types.ReflectionCheck{
		Question:      question,
		InitialAnswer: initialAnswer,
		Concerns:      make([]string, 0),
		Verifications: make([]types.VerificationStep, 0),
		FinalAnswer:   initialAnswer,
		Confidence:    0.5, // Default medium confidence
		WasCorrected:  false,
		CreatedAt:     time.Now(),
	}

	// Step 1: Identify concerns about the answer
	concerns, err := r.identifyConcerns(ctx, question, initialAnswer)
	if err != nil {
		return reflection, fmt.Errorf("failed to identify concerns: %w", err)
	}
	reflection.Concerns = concerns

	if len(concerns) == 0 {
		// No concerns - high confidence
		reflection.Confidence = 0.95
		if r.verbose {
			r.logger.Info("✅ No concerns identified - answer looks good")
		}
		return reflection, nil
	}

	if r.verbose {
		r.logger.Info(fmt.Sprintf("⚠️ Identified %d concern(s):", len(concerns)))
		for i, concern := range concerns {
			r.logger.Info(fmt.Sprintf("   %d. %s", i+1, concern))
		}
	}

	// Step 2: Perform verifications based on concerns
	for _, concern := range concerns {
		var verification types.VerificationStep

		// Determine verification method based on concern type
		if r.needsFactCheck(concern) {
			verification = r.VerifyFacts(ctx, question, initialAnswer)
		} else if r.needsCalculationCheck(concern) {
			verification = r.VerifyCalculation(ctx, question, initialAnswer)
		} else {
			verification = r.CheckConsistency(ctx, question, initialAnswer)
		}

		reflection.Verifications = append(reflection.Verifications, verification)

		if r.verbose {
			status := "✅ PASSED"
			if !verification.Passed {
				status = "❌ FAILED"
			}
			r.logger.Info(fmt.Sprintf("   %s: %s", verification.Method, status))
		}
	}

	// Step 3: Calculate confidence based on verification results
	reflection.Confidence = r.CalculateConfidence(reflection)

	// Step 4: If confidence is low, attempt to correct the answer
	if reflection.Confidence < 0.7 {
		correctedAnswer, err := r.correctAnswer(ctx, question, initialAnswer, reflection)
		if err == nil && correctedAnswer != "" && correctedAnswer != initialAnswer {
			reflection.FinalAnswer = correctedAnswer
			reflection.WasCorrected = true
			if r.verbose {
				r.logger.Info("🔧 Answer was corrected based on verification")
				r.logger.Info(fmt.Sprintf("   Original: %s", initialAnswer))
				r.logger.Info(fmt.Sprintf("   Corrected: %s", correctedAnswer))
			}
		}
	}

	if r.verbose {
		r.logger.Info(fmt.Sprintf("📊 Final confidence: %.2f", reflection.Confidence))
	}

	return reflection, nil
}

// identifyConcerns asks the LLM to identify potential issues with the answer
func (r *Reflector) identifyConcerns(ctx context.Context, question string, answer string) ([]string, error) {
	prompt := fmt.Sprintf(`You are a critical reviewer. Analyze this question and answer pair.
Identify specific concerns about the accuracy or completeness of the answer.

Question: %s
Answer: %s

List any concerns you have about this answer. Consider:
1. Factual accuracy (could the answer be wrong?)
2. Calculation errors (if math is involved)
3. Logical consistency (does it make sense?)
4. Completeness (is anything missing?)

If you have NO concerns and the answer seems correct, respond with: "No concerns identified"

List concerns (one per line):`, question, answer)

	messages := []types.Message{
		{Role: types.RoleUser, Content: prompt},
	}

	response, err := r.provider.Chat(ctx, messages, &types.ChatOptions{
		Temperature: 0.3, // Low temperature for analytical thinking
		MaxTokens:   500,
	})
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(response.Content)
	if content == "" || strings.Contains(strings.ToLower(content), "no concerns") {
		return []string{}, nil
	}

	// Parse concerns (one per line, starting with numbers or dashes)
	lines := strings.Split(content, "\n")
	concerns := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove numbering or bullet points
		line = regexp.MustCompile(`^[\d\.\-\*\)]+\s*`).ReplaceAllString(line, "")
		if line != "" && !strings.Contains(strings.ToLower(line), "no concerns") {
			concerns = append(concerns, line)
		}
	}

	return concerns, nil
}

// needsFactCheck determines if a concern requires fact checking
func (r *Reflector) needsFactCheck(concern string) bool {
	keywords := []string{"fact", "accurate", "correct", "true", "false", "verify", "capital", "date", "year"}
	lowerConcern := strings.ToLower(concern)
	for _, keyword := range keywords {
		if strings.Contains(lowerConcern, keyword) {
			return true
		}
	}
	return false
}

// needsCalculationCheck determines if a concern requires calculation verification
func (r *Reflector) needsCalculationCheck(concern string) bool {
	keywords := []string{"math", "calculation", "number", "compute", "result", "formula", "equation"}
	lowerConcern := strings.ToLower(concern)
	for _, keyword := range keywords {
		if strings.Contains(lowerConcern, keyword) {
			return true
		}
	}
	return false
}

// VerifyFacts uses available tools or LLM to verify factual claims
func (r *Reflector) VerifyFacts(ctx context.Context, question string, answer string) types.VerificationStep {
	step := types.VerificationStep{
		Method: "fact_check",
		Query:  question,
		Passed: false,
	}

	// Try to use web_search or web_fetch tool if available
	if r.registry != nil {
		searchTool := r.registry.Get("web_search")
		if searchTool == nil {
			searchTool = r.registry.Get("web_fetch")
		}

		if searchTool != nil {
			// Extract key fact to verify
			fact := r.extractKeyFact(answer)
			if fact != "" {
				result, err := searchTool.Execute(ctx, map[string]interface{}{
					"query": fact,
					"url":   "", // For web_fetch, leave empty to do search
				})
				if err == nil {
					step.Result = result
					// Simple heuristic: if result contains similar info, fact is verified
					step.Passed = r.resultSupportsAnswer(result, answer)
					return step
				}
			}
		}
	}

	// Fallback: Use LLM to verify based on its knowledge
	prompt := fmt.Sprintf(`Verify if this answer to the question is factually correct.
Only respond with "VERIFIED" if you are highly confident the answer is correct.
Respond with "UNVERIFIED" if you are unsure or the answer is wrong.

Question: %s
Answer: %s

Verification:`, question, answer)

	messages := []types.Message{
		{Role: types.RoleUser, Content: prompt},
	}

	response, err := r.provider.Chat(ctx, messages, &types.ChatOptions{
		Temperature: 0.1, // Very low temperature for factual verification
		MaxTokens:   100,
	})
	if err != nil {
		step.Error = err
		return step
	}

	step.Result = response.Content
	step.Passed = strings.Contains(strings.ToUpper(response.Content), "VERIFIED")
	return step
}

// VerifyCalculation re-computes calculations to verify results
func (r *Reflector) VerifyCalculation(ctx context.Context, question string, answer string) types.VerificationStep {
	step := types.VerificationStep{
		Method: "calculation_verify",
		Query:  question,
		Passed: false,
	}

	// Try to use math_calculate tool if available
	if r.registry != nil {
		mathTool := r.registry.Get("math_calculate")
		if mathTool != nil {
			// Extract mathematical expression from question or answer
			expression := r.extractMathExpression(question, answer)
			if expression != "" {
				result, err := mathTool.Execute(ctx, map[string]interface{}{
					"expression": expression,
				})
				if err == nil {
					step.Result = result
					// Check if calculated result matches answer
					step.Passed = r.calculationMatches(result, answer)
					return step
				}
			}
		}
	}

	// Fallback: Ask LLM to verify calculation
	prompt := fmt.Sprintf(`Verify if the calculation in this answer is correct.
Show your work step by step.
End with "CORRECT" if the calculation is right, "INCORRECT" if wrong.

Question: %s
Answer: %s

Verification:`, question, answer)

	messages := []types.Message{
		{Role: types.RoleUser, Content: prompt},
	}

	response, err := r.provider.Chat(ctx, messages, &types.ChatOptions{
		Temperature: 0.1,
		MaxTokens:   300,
	})
	if err != nil {
		step.Error = err
		return step
	}

	step.Result = response.Content
	step.Passed = strings.Contains(strings.ToUpper(response.Content), "CORRECT")
	return step
}

// CheckConsistency verifies logical consistency with conversation history
func (r *Reflector) CheckConsistency(ctx context.Context, question string, answer string) types.VerificationStep {
	step := types.VerificationStep{
		Method: "consistency_check",
		Query:  question,
		Passed: true, // Default to passed if no history
	}

	// Get recent conversation history
	if r.memory == nil {
		step.Result = "No memory available for consistency check"
		return step
	}

	history, err := r.memory.GetHistory(10)
	if err != nil || len(history) == 0 {
		step.Result = "No conversation history"
		return step
	}

	// Build context from history
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Previous conversation:\n")
	for _, msg := range history {
		contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}

	prompt := fmt.Sprintf(`%s

New question and answer:
Question: %s
Answer: %s

Is this answer logically consistent with the previous conversation?
Consider any contradictions, conflicting facts, or logical inconsistencies.

Respond with "CONSISTENT" if there are no issues.
Respond with "INCONSISTENT" and explain why if you find problems.

Analysis:`, contextBuilder.String(), question, answer)

	messages := []types.Message{
		{Role: types.RoleUser, Content: prompt},
	}

	response, err := r.provider.Chat(ctx, messages, &types.ChatOptions{
		Temperature: 0.2,
		MaxTokens:   200,
	})
	if err != nil {
		step.Error = err
		return step
	}

	step.Result = response.Content
	step.Passed = strings.Contains(strings.ToUpper(response.Content), "CONSISTENT")
	return step
}

// CalculateConfidence computes overall confidence score based on verifications
func (r *Reflector) CalculateConfidence(reflection *types.ReflectionCheck) float64 {
	if len(reflection.Verifications) == 0 {
		// No verifications performed
		if len(reflection.Concerns) == 0 {
			return 0.95 // No concerns, high confidence
		}
		return 0.5 // Had concerns but couldn't verify
	}

	// Count passed vs failed verifications
	passed := 0
	for _, v := range reflection.Verifications {
		if v.Passed {
			passed++
		}
	}

	total := len(reflection.Verifications)
	baseConfidence := float64(passed) / float64(total)

	// Adjust based on number of concerns
	concernsPenalty := 0.05 * float64(len(reflection.Concerns))
	if concernsPenalty > 0.2 {
		concernsPenalty = 0.2 // Cap penalty
	}

	confidence := baseConfidence - concernsPenalty
	if confidence < 0.0 {
		confidence = 0.0
	}
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// correctAnswer attempts to generate a corrected answer based on verification failures
func (r *Reflector) correctAnswer(ctx context.Context, question string, initialAnswer string, reflection *types.ReflectionCheck) (string, error) {
	// Build context of what went wrong
	var issuesBuilder strings.Builder
	issuesBuilder.WriteString("Issues found with the answer:\n")
	for i, v := range reflection.Verifications {
		if !v.Passed {
			issuesBuilder.WriteString(fmt.Sprintf("%d. %s failed: %v\n", i+1, v.Method, v.Result))
		}
	}

	prompt := fmt.Sprintf(`The initial answer to this question had issues.
Please provide a corrected answer.

Question: %s
Initial Answer: %s

%s

Provide a corrected, accurate answer:`, question, initialAnswer, issuesBuilder.String())

	messages := []types.Message{
		{Role: types.RoleUser, Content: prompt},
	}

	response, err := r.provider.Chat(ctx, messages, &types.ChatOptions{
		Temperature: 0.3,
		MaxTokens:   500,
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response.Content), nil
}

// Helper functions

func (r *Reflector) extractKeyFact(answer string) string {
	// Simple extraction: take the first sentence
	sentences := strings.Split(answer, ".")
	if len(sentences) > 0 {
		return strings.TrimSpace(sentences[0])
	}
	return answer
}

func (r *Reflector) resultSupportsAnswer(result interface{}, answer string) bool {
	// Simple heuristic: convert result to string and check for overlap
	resultStr := fmt.Sprintf("%v", result)
	resultLower := strings.ToLower(resultStr)
	answerLower := strings.ToLower(answer)

	// Extract key words from answer (longer than 3 chars)
	words := strings.Fields(answerLower)
	matches := 0
	for _, word := range words {
		if len(word) > 3 && strings.Contains(resultLower, word) {
			matches++
		}
	}

	// If at least 30% of key words match, consider it supporting
	return matches > 0 && float64(matches)/float64(len(words)) > 0.3
}

func (r *Reflector) extractMathExpression(question string, answer string) string {
	// Try to extract numbers and operators from question
	// This is a simple heuristic - could be improved
	re := regexp.MustCompile(`[\d\+\-\*\/\(\)\.\s]+`)
	matches := re.FindAllString(question, -1)
	for _, match := range matches {
		match = strings.TrimSpace(match)
		if len(match) > 3 && (strings.Contains(match, "+") || strings.Contains(match, "*") || strings.Contains(match, "-") || strings.Contains(match, "/")) {
			return match
		}
	}
	return ""
}

func (r *Reflector) calculationMatches(result interface{}, answer string) bool {
	// Extract number from result
	resultStr := fmt.Sprintf("%v", result)

	// Try to parse as JSON first (tool result format)
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(resultStr), &resultMap); err == nil {
		if val, ok := resultMap["result"]; ok {
			resultStr = fmt.Sprintf("%v", val)
		}
	}

	// Extract numbers from both result and answer
	resultNum := r.extractNumber(resultStr)
	answerNum := r.extractNumber(answer)

	if resultNum == "" || answerNum == "" {
		return false
	}

	// Compare numbers (with some tolerance for floating point)
	return resultNum == answerNum || strings.Contains(answer, resultNum)
}

func (r *Reflector) extractNumber(text string) string {
	// Extract first number found
	re := regexp.MustCompile(`[-+]?\d*\.?\d+`)
	match := re.FindString(text)
	return match
}
