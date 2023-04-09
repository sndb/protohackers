#lang racket

(define (start port)
  (define server (tcp-listen port))
  (let loop ()
    (define-values (in out) (tcp-accept server))
    (thread (lambda () (handle in out)))
    (loop)))

(define (handle in out)
  (let loop ()
    (define byte (read-byte in))
    (when (not (eof-object? byte))
      (write-byte byte out)
      (loop)))
  (close-input-port in)
  (close-output-port out))

(start 8080)
