WebSocket Specification
Overview
The Logi backend supports real-time communication with clients (users and admins) through WebSocket connections. Clients can subscribe to various events and receive updates related to bookings, driver statuses, and more.

Connection
URL: ws://localhost:8080/ws

Authentication: Clients must provide a valid JWT token as a query parameter named token.

Example:

bash
Copy code
ws://localhost:8080/ws?token=your_jwt_token_here
Message Structure
All WebSocket messages follow a standardized JSON structure:

json
Copy code
{
  "user_id": "string",     // ID of the user (empty for admin messages)
  "type": "string",        // Type of the message
  "payload": {             // Payload containing relevant data
    // ... message-specific fields ...
  }
}
Message Types
1. new_booking_request
Description: Sent to drivers when a new booking is created and assigned to them.

Payload:

json
Copy code
{
  "booking_id": "booking123",
  "user_id": "user123",
  "pickup_location": {
    "type": "Point",
    "coordinates": [-122.4194, 37.7749]
  },
  "dropoff_location": {
    "type": "Point",
    "coordinates": [-122.4194, 37.7749]
  },
  "vehicle_type": "car",
  "price_estimate": 25.50
}
2. booking_accepted
Description: Sent to the user when a driver accepts their booking.

Payload:

json
Copy code
{
  "booking_id": "booking123",
  "driver_id": "driver456"
}
3. driver_location
Description: Sent to the user to update the current location of the driver during an ongoing trip.

Payload:

json
Copy code
{
  "booking_id": "booking123",
  "latitude": 37.7749,
  "longitude": -122.4194
}
4. status_update
Description: Sent to the user and admin to update the status of the booking.

Payload:

json
Copy code
{
  "booking_id": "booking123",
  "status": "In Transit"
}
5. driver_status_update
Description: Sent to admins to update the status of a driver.

Payload:

json
Copy code
{
  "driver_id": "driver123",
  "status": "Available"
}
Server-to-Client Communication
Admins: Receive messages related to driver status updates (driver_status_update).
Users: Receive messages related to their bookings (new_booking_request, booking_accepted, driver_location, status_update).
Drivers: Receive new_booking_request messages when a new booking is assigned to them.