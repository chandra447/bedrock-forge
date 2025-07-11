import json
import logging
from typing import Dict, Any

# Configure logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)

def lambda_handler(event: Dict[str, Any], context) -> Dict[str, Any]:
    """
    Lambda handler for generating summaries from text.
    
    This function handles the generate_summary function calls from Bedrock agents.
    Expected event structure:
    {
        "messageVersion": "1.0",
        "agent": {...},
        "inputText": "string",
        "sessionId": "string",
        "actionGroup": "string",
        "function": "string",
        "parameters": [...]
    }
    """
    try:
        logger.info(f"Received event: {json.dumps(event, indent=2)}")
        
        # Extract function name and parameters from the Bedrock event
        function_name = event.get('function', '')
        action_group = event.get('actionGroup', '')
        parameters_list = event.get('parameters', [])
        
        # Convert parameters list to dictionary
        parameters = {}
        for param in parameters_list:
            parameters[param.get('name', '')] = param.get('value', '')
        
        logger.info(f"Function: {function_name}, Action Group: {action_group}, Parameters: {parameters}")
        
        if function_name == 'generate_summary':
            return generate_summary(parameters, action_group, function_name)
        else:
            return create_error_response(
                action_group, 
                function_name, 
                f'Unknown function: {function_name}',
                ['generate_summary']
            )
            
    except Exception as e:
        logger.error(f"Error processing request: {str(e)}")
        return create_error_response(
            event.get('actionGroup', ''), 
            event.get('function', ''), 
            'Internal server error',
            str(e)
        )

def generate_summary(parameters: Dict[str, Any], action_group: str, function_name: str) -> Dict[str, Any]:
    """
    Generate summary from text content.
    
    Args:
        parameters: Dictionary containing text, summary_type, and optional max_length
        action_group: Name of the action group that invoked this function
        function_name: Name of the function being called
        
    Returns:
        Dictionary with Bedrock agent response format
    """
    try:
        # Extract parameters
        text = parameters.get('text', '')
        summary_type = parameters.get('summary_type', 'brief')
        max_length = parameters.get('max_length', None)
        
        # Convert max_length to integer if provided
        if max_length is not None:
            try:
                max_length = int(max_length)
            except (ValueError, TypeError):
                max_length = None
        
        # Validate required parameters
        if not text:
            return create_error_response(
                action_group,
                function_name,
                'Missing required parameter: text'
            )
        
        if summary_type not in ['brief', 'detailed', 'bullet_points']:
            return create_error_response(
                action_group,
                function_name,
                'Invalid summary_type',
                {'valid_types': ['brief', 'detailed', 'bullet_points']}
            )
        
        logger.info(f"Generating {summary_type} summary for text of length {len(text)}")
        
        # Generate summary based on type
        if summary_type == 'brief':
            summary = generate_brief_summary(text, max_length)
        elif summary_type == 'detailed':
            summary = generate_detailed_summary(text, max_length)
        elif summary_type == 'bullet_points':
            summary = generate_bullet_points_summary(text, max_length)
        
        # Generate summary ID
        summary_id = f"sum_{hash(text + summary_type) % 100000:05d}"
        
        # Return results using Bedrock response format
        result = {
            'summary_id': summary_id,
            'summary': summary,
            'summary_type': summary_type,
            'original_length': len(text),
            'summary_length': len(summary),
            'compression_ratio': round(len(summary) / len(text), 2) if text else 0
        }
        
        return create_success_response(action_group, function_name, result)
        
    except Exception as e:
        logger.error(f"Error generating summary: {str(e)}")
        return create_error_response(
            action_group,
            function_name,
            'Summary generation failed',
            str(e)
        )

def generate_brief_summary(text: str, max_length: int = None) -> str:
    """Generate a brief summary of the text."""
    # Simple extraction-based summarization
    sentences = [s.strip() for s in text.split('.') if s.strip()]
    
    if not sentences:
        return "No content to summarize."
    
    # Take first and last sentences for brief summary
    if len(sentences) == 1:
        summary = sentences[0]
    elif len(sentences) == 2:
        summary = f"{sentences[0]}. {sentences[1]}"
    else:
        summary = f"{sentences[0]}. {sentences[-1]}"
    
    # Apply max length if specified
    if max_length and len(summary) > max_length:
        summary = summary[:max_length-3] + "..."
    
    return summary

def generate_detailed_summary(text: str, max_length: int = None) -> str:
    """Generate a detailed summary of the text."""
    sentences = [s.strip() for s in text.split('.') if s.strip()]
    
    if not sentences:
        return "No content to summarize."
    
    # Take multiple sentences for detailed summary
    max_sentences = min(5, len(sentences))
    
    if len(sentences) <= max_sentences:
        summary = '. '.join(sentences) + '.'
    else:
        # Take first 3 and last 2 sentences
        first_sentences = sentences[:3]
        last_sentences = sentences[-2:]
        summary = '. '.join(first_sentences + last_sentences) + '.'
    
    # Apply max length if specified
    if max_length and len(summary) > max_length:
        summary = summary[:max_length-3] + "..."
    
    return summary

def generate_bullet_points_summary(text: str, max_length: int = None) -> str:
    """Generate a bullet-point summary of the text."""
    sentences = [s.strip() for s in text.split('.') if s.strip()]
    
    if not sentences:
        return "No content to summarize."
    
    # Convert to bullet points
    max_points = min(5, len(sentences))
    selected_sentences = sentences[:max_points]
    
    # Create bullet points
    bullet_points = []
    for i, sentence in enumerate(selected_sentences):
        # Keep sentences short for bullet points
        if len(sentence) > 100:
            sentence = sentence[:97] + "..."
        bullet_points.append(f"• {sentence}")
    
    summary = '\n'.join(bullet_points)
    
    # Apply max length if specified
    if max_length and len(summary) > max_length:
        # Truncate bullet points to fit
        truncated_points = []
        current_length = 0
        for point in bullet_points:
            if current_length + len(point) + 1 <= max_length - 3:
                truncated_points.append(point)
                current_length += len(point) + 1
            else:
                break
        summary = '\n'.join(truncated_points) + "..."
    
    return summary

def create_success_response(action_group: str, function_name: str, result: Dict[str, Any]) -> Dict[str, Any]:
    """
    Create a successful Bedrock agent response.
    
    Args:
        action_group: Name of the action group
        function_name: Name of the function
        result: The function result data
        
    Returns:
        Properly formatted Bedrock agent response
    """
    return {
        "messageVersion": "1.0",
        "response": {
            "actionGroup": action_group,
            "function": function_name,
            "functionResponse": {
                "responseBody": {
                    "TEXT": {
                        "body": json.dumps(result)
                    }
                }
            }
        }
    }

def create_error_response(action_group: str, function_name: str, error_message: str, details: Any = None) -> Dict[str, Any]:
    """
    Create an error Bedrock agent response.
    
    Args:
        action_group: Name of the action group
        function_name: Name of the function
        error_message: Error message
        details: Additional error details
        
    Returns:
        Properly formatted Bedrock agent error response
    """
    error_data = {
        "error": error_message
    }
    if details:
        error_data["details"] = details
    
    return {
        "messageVersion": "1.0",
        "response": {
            "actionGroup": action_group,
            "function": function_name,
            "functionResponse": {
                "responseState": "FAILURE",
                "responseBody": {
                    "TEXT": {
                        "body": json.dumps(error_data)
                    }
                }
            }
        }
    }
