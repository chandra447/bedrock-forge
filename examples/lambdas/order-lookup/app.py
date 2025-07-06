import json
import os
import logging
from typing import Dict, Any

# Configure logging
log_level = os.getenv('LOG_LEVEL', 'INFO')
logging.basicConfig(level=getattr(logging, log_level))
logger = logging.getLogger(__name__)

def handler(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    Lambda handler for order lookup functionality.
    Called by Bedrock agent action groups.
    """
    try:
        logger.info(f"Received event: {json.dumps(event)}")
        
        # Extract parameters from Bedrock agent
        input_text = event.get('inputText', '')
        parameters = event.get('parameters', [])
        
        # Parse order ID from parameters
        order_id = None
        for param in parameters:
            if param.get('name') == 'order_id':
                order_id = param.get('value')
                break
        
        if not order_id:
            return {
                'statusCode': 400,
                'body': json.dumps({
                    'error': 'Missing required parameter: order_id'
                })
            }
        
        # Mock order lookup (in real implementation, call external API)
        order_data = lookup_order(order_id)
        
        # Return response in Bedrock agent format
        return {
            'statusCode': 200,
            'body': json.dumps({
                'order': order_data,
                'message': f'Successfully retrieved order {order_id}'
            })
        }
        
    except Exception as e:
        logger.error(f"Error processing request: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': 'Internal server error'
            })
        }

def lookup_order(order_id: str) -> Dict[str, Any]:
    """
    Mock function to lookup order details.
    In real implementation, this would call external API.
    """
    # Mock data for demonstration
    mock_orders = {
        '12345': {
            'order_id': '12345',
            'customer_name': 'John Doe',
            'status': 'shipped',
            'items': [
                {'product': 'Widget A', 'quantity': 2, 'price': 29.99},
                {'product': 'Widget B', 'quantity': 1, 'price': 45.50}
            ],
            'total': 105.48,
            'tracking_number': 'TRK789456123'
        },
        '67890': {
            'order_id': '67890',
            'customer_name': 'Jane Smith',
            'status': 'processing',
            'items': [
                {'product': 'Gadget X', 'quantity': 1, 'price': 199.99}
            ],
            'total': 199.99,
            'tracking_number': None
        }
    }
    
    return mock_orders.get(order_id, {
        'order_id': order_id,
        'status': 'not_found',
        'message': 'Order not found in system'
    })