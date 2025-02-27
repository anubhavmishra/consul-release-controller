---
# Default values for consul-canary
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

replicaCount: 1

controller:
  enabled: "true"

  container_config:
    image:
      repository: nicholasjackson/consul-release-controller
      pullPolicy: IfNotPresent
      # Overrides the image tag whose default is the chart appVersion.
      tag: "##VERSION##"

    env:
    - name: HOST_IP
      valueFrom:
        fieldRef:
          fieldPath: status.hostIP
    - name: CONSUL_HTTP_ADDR
      value: https://$(HOST_IP):8501
    - name: CONSUL_CAPATH
      value: /consul/tls/client/ca/tls.crt
    - name: CONSUL_HTTP_TOKEN
      valueFrom: 
        secretKeyRef:
          name: consul-controller-acl-token
          key: token 
      # Additional environment variables to add to the controller
      # deployment
      # - name: MYENV
      #   value: myvalue

    volumeMounts:
      - mountPath: /consul/tls/client/ca
        name: consul-auto-encrypt-ca-cert
      # Additional volume mounts to add to the container can be used to
      # mount configuration or certificates needed by the specific controller implementation.
      # - name: consul-ca
      #   mountPath: /tmp/consul/ca

    resources: {}
      # We usually recommend not to specify default resources and to leave this as a conscious
      # choice for the user. This also increases chances charts run on environments with little
      # resources, such as Minikube. If you do want to specify resources, uncomment the following
      # lines, adjust them as necessary, and remove the curly braces after 'resources:'.
      # limits:
      #   cpu: 100m
      #   memory: 128Mi
      # requests:
      #   cpu: 100m
      #   memory: 128Mi

  additional_volumes: 
    - name: consul-server-ca
      secret:
        secretName: consul-server-cert
    - name: consul-auto-encrypt-ca-cert
      emptyDir:
        medium: Memory
    # Additional volumes to add to the controller pod
    # - name: consul-ca
    #   secret:
    #     secretName: consul-ca-cert

  additional_init_containers:
    - command:
      - /bin/sh
      - -ec
      - |
        consul-k8s get-consul-client-ca \
          -output-file=/consul/tls/client/ca/tls.crt \
          -server-addr=consul-server \
          -server-port=8501 \
          -ca-file=/consul/tls/ca/tls.crt
      image: hashicorp/consul-k8s:0.25.0
      imagePullPolicy: IfNotPresent
      name: get-auto-encrypt-client-ca
      resources:
        limits:
          cpu: 50m
          memory: 50Mi
        requests:
          cpu: 50m
          memory: 50Mi
      volumeMounts:
      - mountPath: /consul/tls/ca
        name: consul-server-ca
      - mountPath: /consul/tls/client/ca
        name: consul-auto-encrypt-ca-cert
    # Add additional init containers to the controller pod
    # - command:
    #   - /bin/sh
    #   - -ec
    #   - |
    #     consul-k8s get-consul-client-ca \
    #       -output-file=/consul/tls/client/ca/tls.crt \
    #       -server-addr=consul-server \
    #       -server-port=8501 \
    #       -ca-file=/consul/tls/ca/tls.crt
    #   image: hashicorp/consul-k8s:0.25.0
    #   imagePullPolicy: IfNotPresent
    #   name: get-auto-encrypt-client-ca

  podAnnotations: {}

  podSecurityContext: {}
    # fsGroup: 2000

  securityContext: {}
    # capabilities:
    #   drop:
    #   - ALL
    # readOnlyRootFilesystem: true
    # runAsNonRoot: true
    # runAsUser: 1000

  autoscaling:
    enabled: false
    minReplicas: 1
    maxReplicas: 100
    targetCPUUtilizationPercentage: 80
    # targetMemoryUtilizationPercentage: 80

  nodeSelector: {}

  tolerations: []

  affinity: {}

webhook:
  enabled: "false"
  type: ClusterIP
  port: 443
  service: consul-release-controller-webhook
  namespaceOverride: ""

  # Allows adding additional DNS Names to the cert generated
  # for the webhook
  additionalDNSNames: []

  failurePolicy: Fail

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  # Specifies whether a service account should be created
  create: true
  # Annotations to add to the service account
  annotations: {}
  # The name of the service account to use.
  # If not set and create is true, a name is generated using the fullname template
  name: ""