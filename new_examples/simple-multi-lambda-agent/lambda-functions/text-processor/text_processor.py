import json
import boto3
import logging
from typing import Dict, Any

# Configure logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Initialize AWS clients
s3_client = boto3.client('s3')

def lambda_handler(event: Dict[str, Any], context) -> Dict[str, Any]:
    """
    Lambda handler for processing text documents from S3.
    
    This function handles the process_text function calls from Bedrock agents.
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
        
        if function_name == 'process_text':
            return process_text(parameters, action_group, function_name)
        else:
            return create_error_response(
                action_group, 
                function_name, 
                f'Unknown function: {function_name}',
                ['process_text']
            )
            
    except Exception as e:
        logger.error(f"Error processing request: {str(e)}")
        return create_error_response(
            event.get('actionGroup', ''), 
            event.get('function', ''), 
            'Internal server error',
            str(e)
        )

def process_text(parameters: Dict[str, Any], action_group: str, function_name: str) -> Dict[str, Any]:
    """
    Process text document from S3 bucket.
    
    Args:
        parameters: Dictionary containing s3_bucket, s3_key, and processing_type
        action_group: Name of the action group that invoked this function
        function_name: Name of the function being called
        
    Returns:
        Dictionary with Bedrock agent response format
    """
    try:
        # Extract parameters
        s3_bucket = parameters.get('s3_bucket')
        s3_key = parameters.get('s3_key')
        processing_type = parameters.get('processing_type', 'extract')
        
        # Validate required parameters
        if not s3_bucket or not s3_key:
            return create_error_response(
                action_group,
                function_name,
                'Missing required parameters',
                {'required': ['s3_bucket', 's3_key']}
            )
        
        # Download file from S3
        logger.info(f"Downloading file from s3://{s3_bucket}/{s3_key}")
        response = s3_client.get_object(Bucket=s3_bucket, Key=s3_key)
        file_content = response['Body'].read()
        
        # Determine file type and process accordingly
        if s3_key.lower().endswith('.txt'):
            text_content = file_content.decode('utf-8')
        elif s3_key.lower().endswith('.json'):
            json_content = json.loads(file_content.decode('utf-8'))
            text_content = json.dumps(json_content, indent=2)
        else:
            # For other file types, treat as text
            text_content = file_content.decode('utf-8', errors='ignore')
        
        # Process based on processing type
        if processing_type == 'extract':
            processed_text = text_content
            metadata = {
                'file_size': len(file_content),
                'character_count': len(text_content),
                'word_count': len(text_content.split()),
                'file_type': s3_key.split('.')[-1].lower()
            }
        elif processing_type == 'analyze':
            processed_text = text_content
            metadata = {
                'file_size': len(file_content),
                'character_count': len(text_content),
                'word_count': len(text_content.split()),
                'line_count': len(text_content.split('\n')),
                'file_type': s3_key.split('.')[-1].lower(),
                'sentiment': 'neutral',  # Placeholder - would use actual sentiment analysis
                'key_topics': ['document', 'text', 'processing']  # Placeholder
            }
        elif processing_type == 'clean':
            # Basic text cleaning
            processed_text = text_content.strip()
            processed_text = '\n'.join(line.strip() for line in processed_text.split('\n') if line.strip())
            metadata = {
                'file_size': len(file_content),
                'original_character_count': len(text_content),
                'cleaned_character_count': len(processed_text),
                'file_type': s3_key.split('.')[-1].lower()
            }
        else:
            return create_error_response(
                action_group,
                function_name,
                'Invalid processing type',
                {'valid_types': ['extract', 'analyze', 'clean']}
            )
        
        # Generate processing ID
        processing_id = f"proc_{hash(s3_bucket + s3_key + processing_type) % 100000:05d}"
        
        # Return results using Bedrock response format
        result = {
            'processing_id': processing_id,
            'status': 'completed',
            'extracted_text': processed_text[:1000],  # Truncate for response
            'metadata': metadata,
            'processing_type': processing_type,
            'source': {
                'bucket': s3_bucket,
                'key': s3_key
            }
        }
        
        return create_success_response(action_group, function_name, result)
        
    except Exception as e:
        logger.error(f"Error processing text: {str(e)}")
        return create_error_response(
            action_group,
            function_name,
            'Text processing failed',
            str(e)
        )

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
