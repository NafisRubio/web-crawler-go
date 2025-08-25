package http

import (
	"encoding/json"
	"net/http"
	"web-crawler-go/internal/core/ports"
)

const ContentType = "Content-Type"
const ApplicationJsonCharsetUtf8 = "application/json; charset=utf-8"

// RespondSuccess writes a JSON response with the given status code and payload.
// It sets the Content-Type header to application/json.
// If encoding the payload fails, it logs the error but cannot send a different
// HTTP status code because the header has already been written.
//
// Parameters:
//
//	w: The http.ResponseWriter to write the response to.
//	logger: A logger for logging potential encoding errors.
//	statusCode: The HTTP status code to send (e.g., http.StatusOK, http.StatusCreated).
//	message: The HTTP status code to send (e.g., http.StatusOK, http.StatusCreated).
//	data: The data to be encoded as JSON. If nil, encodes JSON `null`.
//	         For status 204 No Content, the payload is ignored.
//	pagination: The HTTP status code to send (e.g., http.StatusOK, http.StatusCreated).
func RespondSuccess(w http.ResponseWriter, logger ports.Logger, statusCode int, message string, data interface{}, pagination *Pagination) {
	// Special case: 204 No Content should not have a body or Content-Type
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return
	}

	// Construct the standard API response structure
	apiResponse := Response{
		Status:     "success", // Hardcoded for this success helper
		Message:    message,
		Data:       data, // The actual data payload
		Pagination: pagination,
	}

	// For other statuses, encode the payload to JSON
	responseBytes, err := json.Marshal(apiResponse)
	if err != nil {
		// Log the internal error, but we can't change the response now
		logger.Error("ERROR: Failed to marshal JSON response payload", "error", err)
		// Send a generic 500 Internal Server Error *before* trying to marshal?
		// No, because we want to send the intended status first. If marshalling fails,
		// it's an internal error, but the client might get a partial response or just
		// the headers. Logging is the main action here.
		// We could potentially write a plain text error *after* the headers,
		// but that's usually not ideal for JSON APIs.
		// Consider sending a 500 status code *if* the error occurs *before* WriteHeader,
		// but here we commit to the status first.
		w.Header().Set(ContentType, ApplicationJsonCharsetUtf8) // Set header even for error
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr := w.Write([]byte(`{"status": "error", "message": "Internal server error preparing response"}`))
		if writeErr != nil {
			logger.Error("ERROR: Failed to write 500 error response after marshalling failure", "error", writeErr)
		}
		return // Exit after logging the marshal error
	}

	// If marshalling succeeded:
	w.Header().Set(ContentType, ApplicationJsonCharsetUtf8)
	w.WriteHeader(statusCode)
	_, err = w.Write(responseBytes)
	if err != nil {
		// Log error encountered during write (e.g., connection closed)
		logger.Error("ERROR: Failed to write JSON response body", "error", err)
	}
}

// RespondError writes a JSON error response with the given status code and message.
// It sets the Content-Type header to application/json.
// If encoding the payload fails, it logs the error but cannot send a different
// HTTP status code because the header has already been written.
//
// Parameters:
//
//	w: The http.ResponseWriter to write the response to.
//	logger: A logger for logging potential encoding errors.
//	statusCode: The HTTP status code to send (e.g., http.StatusBadRequest, http.StatusInternalServerError).
//	message: The error message to include in the response.
//	data: Optional error data to be encoded as JSON. If nil, encodes JSON `null`.
func RespondError(w http.ResponseWriter, logger ports.Logger, statusCode int, message string, data interface{}) {
	// Construct the standard API error response structure
	apiResponse := Response{
		Status:  "error", // Hardcoded for this error helper
		Message: message,
		Data:    data, // Optional error data payload
		// Note: Pagination is omitted for error responses as they typically don't need pagination
	}

	// Encode the payload to JSON
	responseBytes, err := json.Marshal(apiResponse)
	if err != nil {
		// Log the internal error, but we can't change the response now
		logger.Error("ERROR: Failed to marshal JSON error response payload", "error", err)
		// Send a generic 500 Internal Server Error with a fallback response
		w.Header().Set(ContentType, ApplicationJsonCharsetUtf8) // Set header even for error
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr := w.Write([]byte(`{"status": "error", "message": "Internal server error preparing error response"}`))
		if writeErr != nil {
			logger.Error("ERROR: Failed to write 500 error response after marshalling failure", "error", writeErr)
		}
		return // Exit after logging the marshal error
	}

	// If marshalling succeeded:
	w.Header().Set(ContentType, ApplicationJsonCharsetUtf8)
	w.WriteHeader(statusCode)
	_, err = w.Write(responseBytes)
	if err != nil {
		// Log error encountered during write (e.g., connection closed)
		logger.Error("ERROR: Failed to write JSON error response body", "error", err)
	}
}
