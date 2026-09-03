/**
 * This is the database schema for testing MySQL support of ozzo-dbx.
 * The following database setup is required in order to run the test:
 * - host: 127.0.0.1
 * - user: travis
 * - pass: <none>
 * - database: ozzo_dbx_test
 */

DROP TABLE IF EXISTS `order_item` CASCADE;
DROP TABLE IF EXISTS `item` CASCADE;
DROP TABLE IF EXISTS `order` CASCADE;
DROP TABLE IF EXISTS `customer` CASCADE;
DROP TABLE IF EXISTS `user` CASCADE;

CREATE TABLE `customer` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `email` varchar(128) NOT NULL,
  `name` varchar(128),
  `address` text,
  `status` int (11) DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `email` varchar(128) NOT NULL,
  `created` date,
  `updated` date,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `item` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `order` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `customer_id` int(11) NOT NULL,
  `created_at` int(11) NOT NULL,
  `total` decimal(10,0) NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `FK_order_customer_id` FOREIGN KEY (`customer_id`) REFERENCES `customer` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `order_item` (
  `order_id` int(11) NOT NULL,
  `item_id` int(11) NOT NULL,
  `quantity` int(11) NOT NULL,
  `subtotal` decimal(10,0) NOT NULL,
  PRIMARY KEY (`order_id`,`item_id`),
  KEY `FK_order_item_item_id` (`item_id`),
  CONSTRAINT `FK_order_item_order_id` FOREIGN KEY (`order_id`) REFERENCES `order` (`id`) ON DELETE CASCADE,
  CONSTRAINT `FK_order_item_item_id` FOREIGN KEY (`item_id`) REFERENCES `item` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `customer` (email, name, address, status) VALUES ('user1@example.com', 'user1', 'address1', 1);
INSERT INTO `customer` (email, name, address, status) VALUES ('user2@example.com', 'user2', NULL, 1);
INSERT INTO `customer` (email, name, address, status) VALUES ('user3@example.com', 'user3', 'address3', 2);

INSERT INTO `user` (email, created) VALUES ('user1@example.com', '2015-01-02');
INSERT INTO `user` (email, created) VALUES ('user2@example.com', now());

INSERT INTO `item` (name) VALUES ('The Go Programming Language');
INSERT INTO `item` (name) VALUES ('Go in Action');
INSERT INTO `item` (name) VALUES ('Go Programming Blueprints');
INSERT INTO `item` (name) VALUES ('Building Microservices');
INSERT INTO `item` (name) VALUES ('Go Web Programming');

INSERT INTO `order` (customer_id, created_at, total) VALUES (1, 1325282384, 110.0);
INSERT INTO `order` (customer_id, created_at, total) VALUES (2, 1325334482, 33.0);
INSERT INTO `order` (customer_id, created_at, total) VALUES (2, 1325502201, 40.0);

INSERT INTO `order_item` (order_id, item_id, quantity, subtotal) VALUES (1, 1, 1, 30.0);
INSERT INTO `order_item` (order_id, item_id, quantity, subtotal) VALUES (1, 2, 2, 40.0);
INSERT INTO `order_item` (order_id, item_id, quantity, subtotal) VALUES (2, 4, 1, 10.0);
INSERT INTO `order_item` (order_id, item_id, quantity, subtotal) VALUES (2, 5, 1, 15.0);
INSERT INTO `order_item` (order_id, item_id, quantity, subtotal) VALUES (2, 3, 1, 8.0);
INSERT INTO `order_item` (order_id, item_id, quantity, subtotal) VALUES (3, 2, 1, 40.0);
