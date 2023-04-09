#lang racket

(require json
         math/number-theory)

(define (start port)
  (define cust (make-custodian))
  (with-handlers ([exn:break? void])
    (parameterize ([current-custodian cust])
      (define listener (tcp-listen port 32 #t))
      (let loop ()
        (accept-and-handle listener)
        (loop))))
  (custodian-shutdown-all cust))

(define (accept-and-handle listener)
  (define cust (make-custodian))
  (parameterize ([current-custodian cust])
    (define-values (in out) (tcp-accept listener))
    (thread
     (lambda ()
       (handle in out)
       (close-input-port in)
       (close-output-port out))))
  (thread
   (lambda ()
     (sleep 20)
     (custodian-shutdown-all cust))))

(define (handle in out)
  (let loop ()
    (match (read-line in)
      [(? eof-object?)
       (void)]
      [line
       (match (response line)
         [#f
          (displayln "malformed" out)]
         [resp
          (displayln resp out)
          (flush-output out)
          (loop)])])))

(define (response req)
  (match (with-handlers ([exn:fail? void]) (string->jsexpr req))
    [(hash-table ('method "isPrime")
                 ('number (? number? n)))
     (jsexpr->string (hash 'method "isPrime" 'prime (and (positive-integer? n) (prime? n))))]
    [_ #f]))

(module+ main
  (start 8080))
